package tools

// 数组到字符串： strings.Replace(strings.Trim(fmt.Sprint(ts.RoomIds), "[]"), " ", ",", -1)

import (
	crand "crypto/rand"
	"fmt"
	"math"
	mrand "math/rand"
	"time"
)

var GRand = mrand.New(mrand.NewSource(time.Now().Unix()))

// gen random string. 问题：数字部分过多
func RandStr(length int) string {
	if length == 0 {
		return ""
	}

	newLen := math.Ceil(float64(length) / 2)
	buf := make([]byte, int(newLen))
	_, err := crand.Read(buf)
	if err != nil {
		fmt.Printf("gen rand str err: %s\n", err.Error())
		return ""
	}

	out := fmt.Sprintf("%x", buf)
	return out[:int(length)]
}

const LetterArray = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
const LetterArrayLen = len(LetterArray)

func GenRandomStr(length int) string {
	bytes := make([]byte, length)
	for i := 0; i < length; i++ {
		b := GRand.Intn(LetterArrayLen)
		bytes[i] = LetterArray[b]
	}

	return string(bytes)
}
