package data

import (
	"database/sql"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stella-go/siu/t/n"
)

var (
	db *sql.DB
)

type TbStudents struct {
	Id         *n.Int    `form:"id" json:"id,omitempty" @free:"table='tb_students',column='id',primary,auto-incrment"`
	No         *n.String `form:"no" json:"no,omitempty" @free:"table='tb_students',column='no'"`
	Name       *n.String `form:"name" json:"name,omitempty" @free:"table='tb_students',column='name'"`
	Age        *n.Int    `form:"age" json:"age,omitempty" @free:"table='tb_students',column='age'"`
	Gender     *n.String `form:"gender" json:"gender,omitempty" @free:"table='tb_students',column='gender'"`
	CreateTime *n.Time   `form:"create_time" json:"create_time,omitempty" @free:"table='tb_students',column='create_time',current-timestamp,round='s'"`
	UpdateTime *n.Time   `form:"update_time" json:"update_time,omitempty" @free:"table='tb_students',column='update_time',current-timestamp,round='s'"`
}

func TestMain(m *testing.M) {
	dsn := "root:root@tcp(127.0.0.1:3306)/test?parseTime=true&collation=utf8_bin&charset=utf8"
	c, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	db = c
	_, err = db.Exec(`
CREATE TABLE IF NOT EXISTS tb_students (
    id INT NOT NULL AUTO_INCREMENT COMMENT 'ROW ID',
    no VARCHAR (32) COMMENT 'STUDENT NUMBER',
    name VARCHAR (64) COMMENT 'STUDENT NAME',
    age INT DEFAULT 1 COMMENT 'STUDENT AGE',
    gender VARCHAR (1) DEFAULT NULL COMMENT 'STUDENT GENDER',
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'CREATE TIME',
    update_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'UPDATE TIME',
    PRIMARY KEY (id)
) ENGINE = INNODB AUTO_INCREMENT = 1 DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = 'STUDENT RECORDS';
	`)
	if err != nil {
		panic(err)
	}
	m.Run()
	db.Close()
}

func TestCreate(t *testing.T) {
	{
		s := &TbStudents{}
		_, err := Create(db, s)
		if err != nil {
			t.Fatal(err)
		}
	}
	{
		s := &TbStudents{Age: NullInt}
		_, err := Create(db, s)
		if err != nil {
			t.Fatal(err)
		}
	}
}
func TestUpdate(t *testing.T) {
	s := &TbStudents{Id: &n.Int{Val: 2}, Age: NullInt}
	_, err := Update(db, s)
	if err != nil {
		t.Fatal(err)
	}
}
