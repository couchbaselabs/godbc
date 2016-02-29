package godbc

import (
	"testing"
)

func TestPing(t *testing.T) {
	db, err := Open("n1ql", "http://localhost:8093")
	if err != nil {
		t.Error("Failed to open.", err.Error())
	}

	// Pinging.
	err = db.Ping()
	if err != nil {
		t.Error("Ping failed.", err.Error())
	}
}

func TestQuery(t *testing.T) {
	db, err := Open("n1ql", "http://localhost:8093")
	if err != nil {
		t.Error("Failed to open.", err.Error())
	}
	// Querying.
	rows, err := db.Query("select * from `beer-sample` where city = 'San Francisco'")
	if err != nil {
		t.Error("Query failed.", err.Error())
	}
	defer rows.Close()

	totalRows := 0
	for rows.Next() {
		totalRows++
	}

	if totalRows != 10 {
		t.Errorf("Query returned %d rows", totalRows)
	}
}

func checkSimpleFields(rows Rows, abvExp float64, nameExp string, isTwentyExp bool, t *testing.T) {
	hasNext := rows.Next()
	if !hasNext {
		t.Error("Unexpected end of rows")
		return
	}

	var abv float64
	var name string
	var is_twenty bool
	err := rows.Scan(&abv, &is_twenty, &name)
	if err != nil {
		t.Error("Scan failed.", err)
	}
	if abv != abvExp {
		t.Error("Expected abv %v, got %v.", abvExp, abv)
	}
	if name != nameExp {
		t.Error("Expected name %v, got %v.", nameExp, name)
	}
	if is_twenty != isTwentyExp {
		t.Error("Expected is_twenty %v, got %v.", isTwentyExp, is_twenty)
	}
}

func TestSimpleFieldValues(t *testing.T) {
	db, err := Open("n1ql", "http://localhost:8093")
	if err != nil {
		t.Error("Failed to open.", err.Error())
	}

	rows, err := db.Query("select name, abv, (ibu = 20) as is_twenty from `beer-sample` where category = 'Other Lager' order by name")
	if err != nil {
		t.Error("Query failed.", err.Error())
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		t.Error("Columns failed.", err.Error())
	}
	var expectedCols = []string{"abv", "is_twenty", "name"}
	if len(cols) != 3 || cols[0] != expectedCols[0] || cols[1] != expectedCols[1] ||
		cols[2] != expectedCols[2] {
		t.Errorf("Expected columns %s, got %s.", expectedCols, cols)
	}

	//
	// Known issue: string fields appear with quotes around them.
	//
	checkSimpleFields(rows, 7.0, "\"Baltika 6\"", true, t)
	checkSimpleFields(rows, 4.8, "\"Estrella Levante Clasica\"", false, t)
	checkSimpleFields(rows, 5.4, "\"Estrella Levante Especial\"", false, t)
	checkSimpleFields(rows, 1.0, "\"Estrella Levante Sin 0.0% Alcohol\"", false, t)

	hasMore := rows.Next()
	if hasMore {
		t.Error("Found more than 4 rows.")
	}
}

/* This test should work, but it doesn't. The driver is unable to parse the result
func TestCount(t *testing.T) {
	db, err := Open("n1ql", "http://localhost:8093")
	if err != nil {
		t.Error("Failed to open.", err.Error())
	}

	rows, err := db.Query("select count(*) as c from `beer-sample` where type = 'beer' and style = 'American-Style Lager' and abv = 5.0")
	if err != nil {
		t.Error("Failed to query.", err.Error())
	}
	rows.Next();
	var c float64
	err = rows.Scan(&c)
	if err != nil {
		t.Error("Failed to scan first result.", err.Error())
	}
	if c != 38.0 {
		t.Errorf("Wrong result on first query. Expected 38, got %v.", c)
	}
}
*/

func TestPrepare(t *testing.T) {
	db, err := Open("n1ql", "http://localhost:8093")
	if err != nil {
		t.Error("Failed to open.", err.Error())
	}

	// Prepare
	stmt, err := db.Prepare("select abv, name from `beer-sample` where type = 'beer' and style = ? and abv > ? order by name")
	if err != nil {
		t.Error("Failed to prepare.", err.Error())
	}

	// First query.
	rows, err := stmt.Query("American-Style Lager", 5.0)
	if err != nil {
		t.Error("Failed to query.", err.Error())
	}
	rows.Next()
	var abv float64
	var name string
	expectedAbv := 5.5
	expectedName := "\"Amber Weizen\"" // Another case of strings being left with extraneous quotes.
	err = rows.Scan(&abv, &name)
	if err != nil {
		t.Error("Failed to scan first result.", err.Error())
	}
	if name != expectedName {
		t.Errorf("Wrong name on first query. Expected %v, got %v.", expectedName, name)
	}
	if abv != expectedAbv {
		t.Errorf("Wrong abv on first query. Expected %v, got %v.", expectedAbv, abv)
	}

	// Second query.
	rows, err = stmt.Query("Porter", 6.0)
	if err != nil {
		t.Error("Failed to query.", err.Error())
	}
	rows.Next()
	expectedAbv = 6.8
	expectedName = "\"(512) Pecan Porter\"" // Another case of strings being left with extraneous quotes.
	err = rows.Scan(&abv, &name)
	if err != nil {
		t.Error("Failed to scan second result.", err.Error())
	}
	if name != expectedName {
		t.Errorf("Wrong name on second query. Expected %v, got %v.", expectedName, name)
	}
	if abv != expectedAbv {
		t.Errorf("Wrong abv on second query. Expected %v, got %v.", expectedAbv, abv)
	}
}

func TestComplex(t *testing.T) {
	db, err := Open("n1ql", "http://localhost:8093")
	if err != nil {
		t.Error("Failed to open.", err.Error())
	}

	rows, err := db.Query("select [1, 2, 3] as fa, { 'a': 'b', 'c': 'd'} as fb")
	if err != nil {
		t.Error("Query failed.", err.Error())
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		t.Error("Columns failed.", err.Error())
	}
	var expectedCols = []string{"fa", "fb"}
	if len(cols) != 2 || cols[0] != expectedCols[0] || cols[1] != expectedCols[1] {
		t.Errorf("Expected columns %s, got %s.", expectedCols, cols)
	}

	//
	// Known issue: It should be possible to scan complex values into slices
	// and maps. Currently, they can only be scanned into strings.
	//
	// arrTarget := make([]uint8, 1)
	// objTarget := make(map[string]string)
	//
	var arrTarget string
	var objTarget string

	rows.Next()
	err = rows.Scan(&arrTarget, &objTarget)
	if arrTarget != "[1,2,3]" {
		t.Errorf("Unexpected array value %v", arrTarget)
	}
	if objTarget != "{\"a\":\"b\",\"c\":\"d\"}" {
		t.Errorf("Unexpected object value %v", objTarget)
	}
}
