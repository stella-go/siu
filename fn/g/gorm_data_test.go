package g

import (
	"fmt"
	"testing"
	"time"

	st "github.com/stella-go/siu/t"
	"github.com/stella-go/siu/t/n"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	db *gorm.DB
)

type TbStudents struct {
	Id         *n.Int    `form:"id" json:"id,omitempty" gorm:"column:id;primarykey;autoIncrement;not null"`
	No         *n.String `form:"no" json:"no,omitempty" gorm:"column:no"`
	Name       *n.String `form:"name" json:"name,omitempty" gorm:"column:name"`
	Age        *n.Int    `form:"age" json:"age,omitempty" gorm:"column:age;default:1"`
	Gender     *n.String `form:"gender" json:"gender,omitempty" gorm:"column:gender;default:NULL"`
	CreateTime *n.Time   `form:"create_time" json:"create_time,omitempty" gorm:"column:create_time;not null;default:current_timestamp"`
	UpdateTime *n.Time   `form:"update_time" json:"update_time,omitempty" gorm:"column:update_time;not null;default:current_timestamp"`
}

func (s *TbStudents) String() string {
	return fmt.Sprintf("TbStudents{Id: %s, No: %s, Name: %s, Age: %s, Gender: %s, CreateTime: %s, UpdateTime: %s}", s.Id, s.No, s.Name, s.Age, s.Gender, s.CreateTime, s.UpdateTime)
}

func TestMain(m *testing.M) {
	dsn := "root:root@tcp(127.0.0.1:3306)/test?parseTime=true&collation=utf8_bin&charset=utf8"
	c, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	db = c
	odb, err := db.DB()
	if err != nil {
		panic(err)
	}
	_, err = odb.Exec(`
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
}

func TestCreate(t *testing.T) {
	{
		s := &TbStudents{}
		err := Create(db, s)
		if err != nil {
			t.Fatal(err)
		}
	}
	{
		s := &TbStudents{Age: st.NullInt}
		err := Create(db, s)
		if err != nil {
			t.Fatal(err)
		}
	}
}
func TestUpdate(t *testing.T) {
	s := &TbStudents{Id: &n.Int{Val: 2}, Age: st.NullInt}
	_, err := Update(db, s)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUpdate2(t *testing.T) {
	s := &TbStudents{Id: &n.Int{Val: 1}, Name: &n.String{Val: "Tom"}, CreateTime: &n.Time{Val: time.Now()}}
	_, err := Update2(db, s)
	if err != nil {
		t.Fatal(err)
	}
}

func TestExec(t *testing.T) {
	{
		stmt := db.Model(&TbStudents{}).Limit(1)
		s, err := QueryExec[TbStudents](stmt)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(s)
	}
	{
		stmt := db.Model(&TbStudents{})
		s, err := QueryExecMany[TbStudents](stmt)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(s)
	}
}
