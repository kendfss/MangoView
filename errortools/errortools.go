package errortools // catlab/errortools

import (
	"log"
)


func Check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func Assert(err error) {
	if err != nil {
		panic(err)
	}
}
