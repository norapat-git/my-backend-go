package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"my-backend-go/config"
)

// TestGetSelect handles GET /api/service/select
// Runs SELECT SYSDATE FROM DUAL as a DB connectivity test.
func TestGetSelect(c *gin.Context) {
	db := config.GetDB()
	rows, err := db.QueryContext(c.Request.Context(), `SELECT sysdate AS SYSDATE FROM dual`)
	if err != nil {
		fmt.Println("TestGetSelect Error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "txt": err.Error()})
		return
	}
	defer rows.Close()

	data, err := scanRowsToMaps(rows)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "txt": err.Error()})
		return
	}
	if len(data) == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "txt": "query send no rows"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "result": data})
}

// TestRevData handles PUT /api/service/send
// Echoes back received data (test endpoint).
func TestRevData(c *gin.Context) {
	var body struct {
		ClientID string      `json:"client_id"`
		Data     interface{} `json:"data"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	if body.ClientID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Data must be isn't null.",
			"err_rec": "",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "received": body})
}

// GetTestPool handles GET /api/service/testpool
// Tests Oracle pool with SELECT SYSDATE.
func GetTestPool(c *gin.Context) {
	db := config.GetDB()
	rows, err := db.QueryContext(c.Request.Context(), `SELECT sysdate AS SYSDATE FROM dual`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	defer rows.Close()

	data, err := scanRowsToMaps(rows)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	if len(data) == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "query no row"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}

// TestInsert handles GET /api/service/testinsert/:x
// Quick insert test (inserts x value to a test record).
func TestInsert(c *gin.Context) {
	x := c.Param("x")
	db := config.GetDB()
	_, err := db.ExecContext(c.Request.Context(),
		`INSERT INTO TABLE_TEST (VAL, CREATED_DATE) VALUES (:1, SYSDATE)`, x)
	if err != nil {
		fmt.Println("TestInsert Error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "inserted", "val": x})
}

// Insertdb handles POST /api/service/insertdb
// Generic insert: receives sql + data from body.
func Insertdb(c *gin.Context) {
	var body struct {
		SQL  string                 `json:"sql"`
		Data map[string]interface{} `json:"data"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.SQL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "sql is required"})
		return
	}

	db := config.GetDB()
	args := mapToArgs(body.Data)
	result, err := db.ExecContext(c.Request.Context(), body.SQL, args...)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "message": err.Error()})
		return
	}
	rowsAffected, _ := result.RowsAffected()
	c.JSON(http.StatusOK, gin.H{"success": rowsAffected > 0, "rowsAffected": rowsAffected})
}

// Deletedb handles DELETE /api/service/deletedb
// Generic delete: receives sql + data from body.
func Deletedb(c *gin.Context) {
	var body struct {
		SQL  string                 `json:"sql"`
		Data map[string]interface{} `json:"data"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.SQL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "sql is required"})
		return
	}

	db := config.GetDB()
	args := mapToArgs(body.Data)
	result, err := db.ExecContext(c.Request.Context(), body.SQL, args...)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "message": err.Error()})
		return
	}
	rowsAffected, _ := result.RowsAffected()
	c.JSON(http.StatusOK, gin.H{"success": rowsAffected > 0, "rowsAffected": rowsAffected})
}

// ClosePool handles GET /api/closepool
// Closes the Oracle DB pool gracefully.
func ClosePool(c *gin.Context) {
	if err := config.CloseDB(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "messege": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "messege": "Disconnect pool."})
}

// scanRowsToMaps converts *sql.Rows into []map[string]interface{}.
func scanRowsToMaps(rows interface{ Columns() ([]string, error); Next() bool; Scan(...interface{}) error; Err() error }) ([]map[string]interface{}, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	result := []map[string]interface{}{}
	for rows.Next() {
		vals := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		row := make(map[string]interface{}, len(cols))
		for i, col := range cols {
			val := vals[i]
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

// mapToArgs converts a map to a positional slice for go-ora.
func mapToArgs(m map[string]interface{}) []interface{} {
	if m == nil {
		return nil
	}
	args := make([]interface{}, 0, len(m))
	for _, v := range m {
		args = append(args, v)
	}
	return args
}
