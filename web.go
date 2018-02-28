package main

/*
 * Basic go web server for any practice or testing.
 */

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/bmizerany/pq"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	defaultPerms = 0664
	host         = "10.254.39.154"
	// host     = "10.24.32.112"
	user     = "postgres"
	dbname   = "macroData"
	password = ""
)

var (
	webPortString string = "8000"
	foodFile      string = "files/food.profiles.txt"
	userFile      string = "files/user.profile.txt"

	db *sql.DB
)

/* Structs for user profile */
type Users struct {
	Users []User `json:"users"`
}

type User struct {
	Id     int    `json:"id"`
	Name   string `json:"name"`
	Goal   string `json:"goal"`
	Macros Macros `json:"macros"`
}

type Macros struct {
	Ratios Ratios `json:"ratios"`
	Grams  Grams  `json:"grams"`
}

type Ratios struct {
	Carbs   int `json:"carbs"`
	Protein int `json:"protein"`
	Fat     int `json:"fat"`
}

type Grams struct {
	Carbs   float64 `json:"carbs"`
	Protein float64 `json:"protein"`
	Fat     float64 `json:"fat"`
}

/* Structs for food info*/
type Foods struct {
	Food []Food `json:"food"`
}

type Food struct {
	Id          int     `json:"id"`
	Name        string  `json:"name"`
	ServingSize float64 `json:"servingSize"`
	Carbs       float64 `json:"carbs"`
	Protein     float64 `json:"protein"`
	Fat         float64 `json:"fat"`
	Sugar       float64 `json::"sugar"`
}

type ProgressBarDatas struct {
	ProgressBarData []ProgressBarData `json:"progressBarData"`
}

type ProgressBarData struct {
	Name 	string `json:"name"`
	Consumed float64 `json:"consumed"`
	Total float64 `json:"total"`
}

//to figure out IP address:
// run ipconfig

// start up the web server
func main() {
	var err error

	// connect to postgres macro db
	psqlInfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", host, user, password, dbname)

	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		fmt.Println("error opening database! ", err)
	}
	defer db.Close()

	// deliver files from the directory
	fileServer := http.FileServer(http.Dir("./"))
	http.Handle("/", fileServer)

	http.HandleFunc("/getUserProfiles", getUserProfiles)
	http.HandleFunc("/getFoodDropdownData", getFoodDropdownData)
	http.HandleFunc("/getUserDailyIntake", getUserDailyIntake)
	http.HandleFunc("/addFoodToUserIntake", addFoodToUserIntake)
	http.HandleFunc("/addNewFood", addFoodToHistory)
	http.HandleFunc("/updateUserConsumedTotals", updateUserConsumedTotals)
	http.HandleFunc("/getUserCurrentConsumption", getUserCurrentConsumption)
	http.HandleFunc("/updateFood", updateFood)
	http.HandleFunc("/getProgressBarData", getProgressBarData)
	http.HandleFunc("/updateProgressBars", updateProgressBars)	

	// register the handler and deliver requests to it
	err = http.ListenAndServe(":"+webPortString, nil)
	if err != nil {
		os.Exit(1)
	}
}

// get user profile from the userprofiles database
func getUserProfiles(w http.ResponseWriter, r *http.Request) {
	tempName := "Erin M"

	// get profile data for current user
	rows, err := db.Query("SELECT * FROM userprofiles where name=$1", tempName)
	if err != nil {
		fmt.Println(err)
	}

	var u User
	for rows.Next() {
		err = rows.Scan(
			&u.Id, &u.Name, &u.Goal,
			&u.Macros.Ratios.Carbs,
			&u.Macros.Ratios.Protein,
			&u.Macros.Ratios.Fat,
			&u.Macros.Grams.Fat,
			&u.Macros.Grams.Protein,
			&u.Macros.Grams.Carbs)
		if err != nil {
			fmt.Println(fmt.Sprintf("error scanning user profile row data: %v", err))
			return
		}
	}

	// marshal the data to send back to the front end
	jsondata, err := json.Marshal(u)
	if err != nil {
		fmt.Println(fmt.Sprintf("unable to marshal user data: %f", err))
	}
	fmt.Fprintf(w, "%s", []byte(jsondata))
}

// getUserCurrentConsumption gets the current daily number of macros consumed
// by the user. In the case that there is not data for the current date a row
// will be created and populated with 0s for each macro
func getUserCurrentConsumption(w http.ResponseWriter, r *http.Request) {
	user_id := r.FormValue("user_id")
	t := time.Now()

	rows, err := db.Query("SELECT COUNT(*) from usermacrototal where date=$1 and user_id=$2", t.Format("2006-01-02"), user_id)
	if err != nil {
		fmt.Println(err)
	}

	var count int
	for rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			fmt.Println("Error scanning macro row count data: ", err)
			return
		}
	}

	var g Grams
	if count == 0 {
		// no macro entry for the current date, create one
		g.Carbs = 0
		g.Protein = 0
		g.Fat = 0
		_, err := db.Exec("INSERT INTO usermacrototal (date, carbs, protein, fat, user_id) VALUES ($1, $2, $3, $4, $5)",
			t.Format("2006-01-02"),
			g.Carbs,
			g.Protein,
			g.Fat,
			user_id)
		if err != nil {
			fmt.Println("error updating macro total table: ", err)
			return
		}
	} else {
		// get totals
		rows, err := db.Query("SELECT carbs, protein, fat from usermacrototal where date=$1 and user_id=$2", t.Format("2006-01-02"), user_id)
		if err != nil {
			fmt.Println("Error querying macro consumption table: ", err)
			return
		}
		for rows.Next() {
			err = rows.Scan(&g.Carbs, &g.Protein, &g.Fat)
			if err != nil {
				fmt.Println("Error scanning macro consumption row data: ", err)
				return
			}
		}
	}

	// marshal the data and return it
	// fmt.Fprintf(w, "macro totals updated successfully (%d row affected)\n", rowsAffected)

	// marshal the data and return
	jsondata, err := json.Marshal(g)
	if err != nil {
		fmt.Println(fmt.Sprintf("unable to marshal macro totals data: %v", err))
	}
	fmt.Fprintf(w, "%s", []byte(jsondata))

}

//select carbs, protein, fat, mg_carbs, mg_protein, mg_fat from usermacrototal join userprofiles up on up.id = 1 where date = '2017-06-13'

func getProgressBarData(w http.ResponseWriter, r *http.Request) { 
	user_id := r.FormValue("user_id")
	t := time.Now()

	var pbd ProgressBarDatas

	// get current macro totals from usermacroconsumption data
	rows, err := db.Query("select carbs, protein, fat, mg_carbs, mg_protein, mg_fat from usermacrototal join userprofiles up on up.id = $1 where date = $2", user_id, t.Format("2006-01-02"))
	if err != nil {
		fmt.Println(err)
	}

	var cpb ProgressBarData 
	cpb.Name = "carbs"
	var ppb ProgressBarData 
	ppb.Name = "protein"
	var fpb ProgressBarData
	fpb.Name = "fat"

	for rows.Next() {
		err = rows.Scan(&cpb.Consumed, &ppb.Consumed, &fpb.Consumed, &cpb.Total, &ppb.Total, &fpb.Total)
		if err != nil {
			fmt.Println("Error scanning macro consumption row data: ", err)
			return
		}
	}	

	pbd.ProgressBarData = append(pbd.ProgressBarData, cpb)
	pbd.ProgressBarData = append(pbd.ProgressBarData, ppb)
	pbd.ProgressBarData = append(pbd.ProgressBarData, fpb)

	fmt.Println(pbd)

	jsondata, err := json.Marshal(pbd)
	if err != nil {
		fmt.Println(fmt.Sprintf("unable to marshal macro totals data: %v", err))
	}
	fmt.Fprintf(w, "%s", []byte(jsondata))
}

func updateProgressBars(w http.ResponseWriter, r *http.Request) {
	user_id := r.FormValue("user_id")
	carbs, err := strconv.ParseFloat(r.FormValue("carbs"), 64)
	protein, err := strconv.ParseFloat(r.FormValue("protein"), 64)
	fat, err := strconv.ParseFloat(r.FormValue("fat"), 64)
	t := time.Now()

	// get current macro totals from usermacroconsumption data
	rows, err := db.Query("SELECT carbs, protein, fat FROM usermacrototal where user_id=$1 and date=$2", user_id, t.Format("2006-01-02"))
	if err != nil {
		fmt.Println(err)
	}

	var g Grams
	for rows.Next() {
		err = rows.Scan(&g.Carbs, &g.Protein, &g.Fat)
		if err != nil {
			fmt.Println("Error scanning macro consumption row data: ", err)
			return
		}
	}

	// update macros and add new totals to the db
	g.Carbs += carbs
	g.Protein += protein
	g.Fat += fat

	_, err = db.Exec("UPDATE usermacrototal SET carbs=$1, protein=$2, fat=$3 WHERE date=$4 and user_id=$5",
		g.Carbs,
		g.Protein,
		g.Fat,
		t.Format("2006-01-02"),
		user_id)
	if err != nil {
		fmt.Println("error updating macro total table: ", err)
		return
	}



	var pbd ProgressBarDatas

	// get current macro totals from usermacroconsumption data
	rows, err = db.Query("select mg_carbs, mg_protein, mg_fat from userprofiles where id = $1", user_id)
	if err != nil {
		fmt.Println(err)
	}

	var cpb ProgressBarData 
	cpb.Name = "carbs"
	cpb.Consumed = g.Carbs
	var ppb ProgressBarData 
	ppb.Name = "protein"
	ppb.Consumed = g.Protein
	var fpb ProgressBarData
	fpb.Name = "fat"
	fpb.Consumed = g.Fat

	for rows.Next() {
		err = rows.Scan(&cpb.Total, &ppb.Total, &fpb.Total)
		if err != nil {
			fmt.Println("Error scanning macro consumption row data: ", err)
			return
		}
	}	

	pbd.ProgressBarData = append(pbd.ProgressBarData, cpb)
	pbd.ProgressBarData = append(pbd.ProgressBarData, ppb)
	pbd.ProgressBarData = append(pbd.ProgressBarData, fpb)


	jsondata, err := json.Marshal(pbd)
	if err != nil {
		fmt.Println(fmt.Sprintf("unable to marshal macro totals data: %v", err))
	}
	fmt.Fprintf(w, "%s", []byte(jsondata))	
}

// updateUserConsumedTotals updates the total number of grams consumed
// when a new food is added to the userdailyintake list
func updateUserConsumedTotals(w http.ResponseWriter, r *http.Request) {
	user_id := r.FormValue("user_id")
	carbs, err := strconv.ParseFloat(r.FormValue("carbs"), 64)
	protein, err := strconv.ParseFloat(r.FormValue("protein"), 64)
	fat, err := strconv.ParseFloat(r.FormValue("fat"), 64)
	t := time.Now()

	// get current macro totals from usermacroconsumption data
	rows, err := db.Query("SELECT carbs, protein, fat FROM usermacrototal where date=$1 and user_id=$2", t.Format("2006-01-02"), user_id)
	if err != nil {
		fmt.Println(err)
	}

	var g Grams
	for rows.Next() {
		err = rows.Scan(&g.Carbs, &g.Protein, &g.Fat)
		if err != nil {
			fmt.Println("Error scanning macro consumption row data: ", err)
			return
		}
	}

	// update macros and add new totals to the db
	g.Carbs += carbs
	g.Protein += protein
	g.Fat += fat

	_, err = db.Exec("UPDATE usermacrototal SET carbs=$1, protein=$2, fat=$3 WHERE date=$4 and user_id=$5",
		g.Carbs,
		g.Protein,
		g.Fat,
		t.Format("2006-01-02"),
		user_id)
	if err != nil {
		fmt.Println("error updating macro total table: ", err)
		return
	}



	// marshal the data and return
	jsondata, err := json.Marshal(g)
	if err != nil {
		fmt.Println(fmt.Sprintf("unable to marshal macro totals data: %v", err))
	}
	fmt.Fprintf(w, "%s", []byte(jsondata))

}

func getFoodDropdownData(w http.ResponseWriter, r *http.Request) {
	var foodList Foods

	// get all from foodprofiles database
	rows, err := db.Query("SELECT * FROM foodprofiles")
	if err != nil {
		fmt.Println(err)
	}

	// add all foods to foodList struct
	for rows.Next() {
		var f Food
		err = rows.Scan(
			&f.Id,
			&f.Name,
			&f.ServingSize,
			&f.Carbs,
			&f.Protein,
			&f.Fat,
			&f.Sugar)
		if err != nil {
			fmt.Println("Error scanning food profile row data: ", err)
			return
		}

		foodList.Food = append(foodList.Food, f)
	}

	fmt.Println(foodList)

	// marshal the data and return
	jsondata, err := json.Marshal(foodList)
	if err != nil {
		fmt.Println(fmt.Sprintf("unable to marshal food dropdown data: %v", err))
	}
	fmt.Fprintf(w, "%s", []byte(jsondata))

}

func getUserDailyIntake(w http.ResponseWriter, r *http.Request) {
	t := time.Now()
	user_id, err := strconv.Atoi(r.FormValue("user_id"))
	if err != nil {
		fmt.Println("error converting user id to int: ", err)
	}
	var foodList Foods
	// intantiate food array to prevent null value when list is empty
	foodList.Food = make([]Food, 0)

	// get all from foodprofiles database
	rows, err := db.Query("select name, servingsize, carbs, protein, fat, sugar from foodprofiles fp join userdailyintake udi ON fp.id = udi.food_id WHERE udi.date = $1 AND udi.user_id = $2", t.Format("2006-01-02"), user_id)
	if err != nil {
		fmt.Println("Error querying user daily intake: ", err)
	}

	for rows.Next() {
		var f Food
		err = rows.Scan(
			// &f.Id,
			&f.Name,
			&f.ServingSize,
			&f.Carbs,
			&f.Protein,
			&f.Fat,
			&f.Sugar)
		if err != nil {
			fmt.Println("Error scanning food profile row data: ", err)
			return
		}

		foodList.Food = append(foodList.Food, f)
	}

	// marshal the data and return
	jsondata, err := json.Marshal(foodList)
	if err != nil {
		fmt.Println(fmt.Sprintf("unable to nmarshal user intake data: %v", err))
	}
	fmt.Fprintf(w, "%s", []byte(jsondata))
}

func updateFood(w http.ResponseWriter, r *http.Request) {
	var f Food
	var err error

	f.Name = r.FormValue("name")
	f.Id, err = strconv.Atoi(r.FormValue("id"))
	f.ServingSize, err = strconv.ParseFloat(r.FormValue("servingSize"), 64)
	f.Carbs, err = strconv.ParseFloat(r.FormValue("carbs"), 64)
	f.Protein, err = strconv.ParseFloat(r.FormValue("protein"), 64)
	f.Fat, err = strconv.ParseFloat(r.FormValue("fat"), 64)

	_, err = db.Exec("UPDATE foodprofiles SET name=$1, servingSize=$2, carbs=$3, protein=$4, fat=$5 WHERE name=$6",
		f.Name,
		f.ServingSize,
		f.Carbs,
		f.Protein,
		f.Fat,
		f.Name)
	if err != nil {
		fmt.Println("Error updating food profile: ", err)
		return
	}


	jf, err := json.Marshal(f)
	fmt.Fprintf(w, "%s", []byte(jf))		
}

func addFoodToUserIntake(w http.ResponseWriter, r *http.Request) {
	var f Food
	var user_id int
	var err error

	f.Name = r.FormValue("name")
	f.Id, err = strconv.Atoi(r.FormValue("id"))
	f.ServingSize, err = strconv.ParseFloat(r.FormValue("servingSize"), 64)
	f.Carbs, err = strconv.ParseFloat(r.FormValue("carbs"), 64)
	f.Protein, err = strconv.ParseFloat(r.FormValue("protein"), 64)
	f.Fat, err = strconv.ParseFloat(r.FormValue("fat"), 64)

	//get user id
	user_id, err = strconv.Atoi(r.FormValue("user_id"))

	if err != nil {
		fmt.Println("error preparing fields to add to database: ", err)
		return
	}

	t := time.Now()

	_, err = db.Exec("INSERT INTO userdailyintake (user_id, food_id, date) VALUES($1, $2, $3)",
		user_id,
		f.Id,
		t.Format("2006-01-02"))
	if err != nil {
		fmt.Println("Error adding food to daily intake database: ", err)
		return
	}

	// rowsAffected, err := result.RowsAffected()
	// if err != nil {
	// 	// http.Error(w, http.StatusText(500), 500)
	// 	return
	// }

	// fmt.Fprintf(w, "Food %s created successfully (%d row affected)\n", f.Name, rowsAffected)

	// return added food to front end to add to dropdown
	jf, err := json.Marshal(f)
	fmt.Fprintf(w, "%s", []byte(jf))
}

func foodAlreadyExists(foodName string) bool {

	rows, err := db.Query("SELECT COUNT(*) from foodprofiles where name=$1", foodName)
	if err != nil {
		fmt.Println(err)
	}

	var count int
	for rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			fmt.Println("Error scanning macro row count data: ", err)
			return false
		}
	}	

	if count == 0 {
		return false
	}

	return true
}

// addFoodToHistory prepares and adds fields to foodprofiles database
// if successful, func will return a json struct of the newly added food
func addFoodToHistory(w http.ResponseWriter, r *http.Request) {
	var f Food
	var err error
	f.Name = r.FormValue("name")
	f.ServingSize, err = strconv.ParseFloat(r.FormValue("servingSize"), 64)
	f.Carbs, err = strconv.ParseFloat(r.FormValue("carbs"), 64)
	f.Protein, err = strconv.ParseFloat(r.FormValue("protein"), 64)
	f.Fat, err = strconv.ParseFloat(r.FormValue("fat"), 64)

	if err != nil {
		fmt.Println("error preparing fields to add to database: ", err)
		return
	}

	if foodAlreadyExists(f.Name) {
		// need to prompt the user to see if they want to update the food
		fmt.Fprintf(w, "%s", []byte("food already exists"))
		return
	}	

	_, err = db.Exec("INSERT INTO foodprofiles (name, servingSize, carbs, protein, fat, sugar) VALUES($1, $2, $3, $4, $5, $6)",
		f.Name,
		f.ServingSize,
		f.Carbs,
		f.Protein,
		f.Fat,
		f.Sugar)
	if err != nil {
		fmt.Println("error adding food to database: ", err)
		return
	}

	// rowsAffected, err := result.RowsAffected()
	// if err != nil {
	// 	// http.Error(w, http.StatusText(500), 500)
	// 	return
	// }

	// fmt.Fprintf(w, "Food %s created successfully (%d row affected)\n", f.Name, rowsAffected)

	// get food id created by the database
	rows, err := db.Query("SELECT id FROM foodprofiles where name=$1", f.Name)
	if err != nil {
		fmt.Println("Error querying datavase for food profile id: ", err)
		return
	}

	// add food id to food struct
	for rows.Next() {
		err = rows.Scan(&f.Id)
		if err != nil {
			fmt.Println("could not scan food id: ", err)
			return
		}
	}

	// return added food to front end to add to dropdown
	jf, err := json.Marshal(f)
	fmt.Fprintf(w, "%s", []byte(jf))
}
