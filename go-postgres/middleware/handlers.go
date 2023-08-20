package middleware

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "github.com/lib/pq"

	"github.com/Valgard/godotenv"
	"github.com/gorilla/mux"
	"github.com/pirateunclejack/go-practice/go-postgres/models"
)

type response struct {
	ID int64 `json:"id,omitempty"`
	Message string `json:"mesage,omitempty"`
}

func CreateConnection()	*sql.DB {
	dotenv := godotenv.New()
	if err := dotenv.Load(".env"); err != nil {
		log.Fatal("Error loading .env file")
		panic(err)
	}

	psqlconn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", os.Getenv("HOST"), os.Getenv("PORT"), os.Getenv("DBUSER"), os.Getenv("PASSWORD"), os.Getenv("DBNAME"), os.Getenv("SSLMODE"))
	db, err := sql.Open("postgres", psqlconn)
	if err != nil {
		log.Fatalf("Error open database, %v", err)
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("Error connecting to database")
		panic(err)
	}

	fmt.Println("Successfully connected to postgres database")
	return db
}

func CreateStock(w http.ResponseWriter, r *http.Request) {
	var stock models.Stock

	err := json.NewDecoder(r.Body).Decode(&stock)
	if err != nil {
		log.Fatalf("unable to decode the request body. %v", err)
	}

	insertID := insertStock(stock)

	res := response{
		ID: insertID,
		Message: "stock created successfully",
	}

	json.NewEncoder(w).Encode(res)
}

func GetStock(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	id, err := strconv.Atoi(params["id"])
	if err != nil {
		log.Fatalf("unable to convert the string into int when getstock %v", err)
	}

	stock, err := getStock(int64(id))
	if err != nil {
		log.Fatalf("unable to get stock %v", err)
	}

	json.NewEncoder(w).Encode(stock)
	
}

func GetAllStock(w http.ResponseWriter, r *http.Request) {
	stocks, err := getAllStocks()
	if err != nil {
		log.Fatalf("unable to get all the stocks: %v", err)
	}

	json.NewEncoder(w).Encode(stocks)
}
func UpdateStock(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	id, err := strconv.Atoi(params["id"])
	if err != nil {
		log.Fatalf("unable to convert the string into int when updatestock %v", err)
	}

	var stock models.Stock

	err = json.NewDecoder(r.Body).Decode(&stock)
	if err != nil {
		log.Fatalf("unable to decode the request body when updatestock %v", err)
	}

	updatedRows := updateStock(int64(id),stock)
	msg := fmt.Sprintf("Stock updated successfully. Total rows/records affected %v", updatedRows)
	res := response {
		ID: int64(id),
		Message: msg,
	}

	json.NewEncoder(w).Encode(res)
}
func DeleteStock(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		log.Fatalf("unable to convert string to int when deletestock. %v", err)
	}

	deletedRows := deleteStock(int64(id))
	msg := fmt.Sprintf("Stock deleted successfully. Total rows/records %v", deletedRows)

	res := response{
		ID: int64(id),
		Message: msg,
	}

	json.NewEncoder(w).Encode(res)
}

func insertStock(stock models.Stock) int64 {
	db := CreateConnection()
	defer db.Close()
	sqlStatement := `INSERT INTO stocks(name, price, company) VALUES ($1, $2, $3) RETURNING stockid`
	var id int64

	err := db.QueryRow(sqlStatement, stock.Name, stock.Price, stock.Company).Scan(&id)

	if err != nil {
		log.Fatalf("unable to execute the query. %v", err)
	}

	fmt.Printf("Inserted a single record %v", id)
	return id
}

func getStock(id int64)(models.Stock, error) {
	db := CreateConnection()
	defer db.Close()

	var stock models.Stock

	sqlStatement := `SELECT * FROM stocks WHERE stockid=$1`

	row := db.QueryRow(sqlStatement, id)

	err := row.Scan(&stock.StockID, &stock.Name, &stock.Price, &stock.Company)
	switch err {
		case sql.ErrNoRows:
			fmt.Println("no rows were returned!")
			return stock, nil
		case nil:
			return stock, nil
		default:
			log.Fatalf("unable to scan the rows. %v", err)
	}

	return stock, err
}

func getAllStocks() ([]models.Stock, error) {
	db := CreateConnection()
	defer db.Close()

	var stocks []models.Stock
	sqlStatement := `SELECT * FROM stocks`

	rows, err := db.Query(sqlStatement)
	if err != nil {
		log.Fatalf("unable to execute the query. %v", err)
	}

	defer rows.Close()

	for rows.Next(){
		var stock models.Stock

		err := rows.Scan(&stock.StockID, &stock.Name, &stock.Price, &stock.Company)
		if err != nil {
			log.Fatalf("Unable to scan the row. %v", err)
		}
		stocks = append(stocks, stock)
	}

	return stocks, nil
}

func updateStock(id int64, stock models.Stock) int64 {
	db := CreateConnection()
	defer db.Close()

	sqlStatement := `UPDATE stocks SET name=$2, price=$3, company=$4 WHERE stockid=$1`
	res, err := db.Exec(sqlStatement, id, stock.Name, stock.Price, stock.Company)
	if err != nil {
		log.Fatalf("unable to execute query %v", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Fatalf("error while checking the affected rows: %v", err)
	}
	fmt.Printf("Total rows/recodes affected %v\n", rowsAffected)
	return rowsAffected

}

func deleteStock(id int64) int64 {
	db := CreateConnection()
	defer db.Close()

	sqlStatement := `DELETE FROM stocks WHERE stockid=$1`
	res, err := db.Exec(sqlStatement, id)
	if err != nil {
		log.Fatalf("unable to execute the query %v", err)
	}
	
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Fatalf("Error while checking affected rows %v", err)
	}

	fmt.Printf("Total rows/records affected: %v", rowsAffected)
	return rowsAffected
}
