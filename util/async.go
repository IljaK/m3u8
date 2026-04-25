package util

import "sync"

func SomeFunc(tableName string, rowId int64) {

}

func SomeFunc2(args ...interface{}) {

}

func SomeFUncLaunch() {

	wg := sync.WaitGroup{}
	AsyncLaunch(&wg, SomeFunc2, "123", 1)
}

func AsyncLaunch(wg *sync.WaitGroup, fn func(args ...interface{}), args ...interface{}) {
	if wg == nil {
		return
	}
	wg.Add(1)
	go asyncRun(wg, fn, args...)
}

func asyncRun(wg *sync.WaitGroup, fn func(args ...interface{}), args ...interface{}) {
	defer wg.Done()
	fn(args...)
}
