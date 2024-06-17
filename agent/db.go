package main

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"langchaingo-learn/util"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	db      *gorm.DB
	db_once sync.Once
)

func createMysqlDB(dbname, host, user, pass string, port int) *gorm.DB {
	// data source name 是 tester:123456@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=True&loc=Local
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, pass, host, port, dbname) //mb4兼容emoji表情符号
	var err error
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{PrepareStmt: true}) //启用PrepareStmt，SQL预编译，提高查询效率
	if err != nil {
		log.Fatalf("connect to mysql use dsn %s failed: %s", dsn, err) //panic() os.Exit(2)
	}
	//设置数据库连接池参数，提高并发性能
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(100) //设置数据库连接池最大连接数
	sqlDB.SetMaxIdleConns(20)  //连接池最大允许的空闲连接数，如果没有sql任务需要执行的连接数大于20，超过的连接会被连接池关闭。
	log.Printf("connect to mysql db %s", dbname)
	return db
}

func GetDBConnection() *gorm.DB { //单例
	if db == nil {
		db_once.Do(func() {
			dbName := "test"
			viper := util.CreateConfig("mysql")
			host := viper.GetString(dbName + ".host")
			port := viper.GetInt(dbName + ".port")
			user := viper.GetString(dbName + ".user")
			pass := viper.GetString(dbName + ".pass")
			db = createMysqlDB(dbName, host, user, pass, port)
		})
	}

	return db
}

type Student struct {
	Name  string
	Score float64
	City  string
}

func (Student) TableName() string {
	return "student"
}

func GetScoreOfStudent(name string) (float64, error) {
	db := GetDBConnection()
	var student Student
	// defer func() {
	// 	fmt.Printf("%s的成绩是%.1f\n", name, student.Score)
	// }()
	err := db.Select("score").Where("name=?", name).First(&student).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Println(err)
			return 0.0, err
		}
		return 0.0, nil
	} else {
		return student.Score, nil
	}
}

func GetCityOfStudent(name string) (string, error) {
	db := GetDBConnection()
	var student Student
	defer func() {
		fmt.Printf("%s的城市是%s\n", name, student.City)
	}()
	err := db.Select("city").Where("name=?", name).First(&student).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Println(err)
			return "", err
		}
		return "", nil
	} else {
		return student.City, nil
	}
}

// func main() {
// 	fmt.Println(GetCityOfStudent("Tom"))
// }
