package random

import (
	_ "embed"
	"time"

	// ランダムなデータを生成
	"github.com/brianvoe/gofakeit/v7"
	// 日本語の名前や地名のダミーデータを生成
	"github.com/mattn/go-gimei"
)

var (
	dateStart = time.Date(1954, 1, 1, 0, 0, 0, 0, time.UTC)   // 70歳ぐらい
	dateEnd   = time.Date(2006, 12, 31, 0, 0, 0, 0, time.UTC) // 18歳ぐらい
)

func init() {
	// 内部データをロードさせておく
	_ = gimei.NewName()
}

func GenerateLastName() string    { return gimei.NewName().Last.Kanji() }
func GenerateFirstName() string   { return gimei.NewName().First.Kanji() }
func GenerateUserName() string    { return gofakeit.Username() }
func GenerateDateOfBirth() string { return gofakeit.DateRange(dateStart, dateEnd).Format("2006-01-02") }
