package main

import (
	"fmt"
	"sync"
	"time"
)

// A Task represents a meaninglessness of our life
type Task struct {
	id         int
	createTime string // время создания
	finishTime string // время выполнения
	taskResult string // !!! возможно сделать стринг
}

type ResultTasks struct {
	DoneTask  []Task
	ErrorTask []Task
}

var (
	generateTime        time.Duration = 10
	outputTime          time.Duration = 4
	generationfrequency time.Duration = 1000
)

func taskGenerate(superChan chan Task, wg *sync.WaitGroup) { // !!!функция только на отдавания в канал// !!! а переименновать
	//сделать тайминг работы 10
	start := time.Now()
	// fmt.Println("taskGenerate")
	for {

		if time.Since(start) >= generateTime*time.Second {

			close(superChan) // Закрываем канал после 10 секунд
			// fmt.Println("Generate - stop")

			return // Выходим из функции
		} // !!!

		createTime := time.Now().Format(time.RFC3339)
		if time.Now().Nanosecond()%2 > 0 { // вот такое условие появления ошибочных тасков !!!
			createTime = "Some error occured"
			// fmt.Println("_____1")
		}
		fmt.Println("_____1.1")
		wg.Add(1)
		superChan <- Task{createTime: createTime, id: int(time.Now().Unix())} // передаем таск на выполнение
		time.Sleep(generationfrequency * time.Millisecond)                    // Ограничение на 1 секунду

	}
	// !!! возможно здесь надо сделать слип 10 секунд, и счетчик горутин
}

func tickerResult(mu *sync.Mutex, wg *sync.WaitGroup, resultTasks *ResultTasks, resultChan chan ResultTasks, stop chan bool) {
	ticker := time.NewTicker(outputTime * time.Second)

	for range ticker.C {

		mu.Lock()
		resultChan <- *resultTasks
		mu.Unlock()

		// _, ok := <-stop
		// if !ok {
		// 	fmt.Println("____")
		// 	// wg.Done()
		// 	close(resultChan)
		// 	return
		// }

	}

}

func printResult(resultChan chan ResultTasks, wg *sync.WaitGroup) {
	defer wg.Done()
	for result := range resultChan {
		println("Done tasks:")
		for _, task_d := range result.DoneTask {
			fmt.Println(task_d)
			// fmt.Println("task_r")
		}
		println("Errors:")
		for _, task_e := range result.ErrorTask {
			fmt.Printf("Task id %d time %s, error %s\n", task_e.id, task_e.createTime, task_e.taskResult)
			// fmt.Println(task_r)
		}
	}

}

func taskWorker(task Task, doneChan chan Task, mu *sync.Mutex, wg *sync.WaitGroup, result *ResultTasks) { // !!! не нужно чтоб возвращал таск
	defer wg.Done()
	// fmt.Println("k__")
	// fmt.Println("taskWorker")
	// undoneTasks := make(chan error)
	tt, _ := time.Parse(time.RFC3339, task.createTime)

	if tt.After(time.Now().Add(-20 * time.Second)) {
		// fmt.Println("k2__1")
		task.taskResult = "task has been successed"
		task.finishTime = time.Now().Format(time.RFC3339Nano)
		// doneChan <- task
		// fmt.Println("t__")
		mu.Lock()
		// fmt.Println("t__k")
		result.DoneTask = append(result.DoneTask, task) // !!! rпереименновать
		mu.Unlock()
	} else {
		task.taskResult = "something went wrong"
		task.finishTime = time.Now().Format(time.RFC3339Nano)
		// undoneTasks <- fmt.Errorf("Task id %d time %s, error %s", task.id, task.createTime, task.taskResult)
		// doneChan <- task
		mu.Lock()
		result.ErrorTask = append(result.ErrorTask, task) // !!! rпереименновать
		mu.Unlock()
		// fmt.Println("t2__")
	}

	// fmt.Println("k2__")
	time.Sleep(time.Millisecond * 150) //!!! не очень понятно зачем эта задержка, функциональной нагрузке не несет
}

func main() {
	start := time.Now()
	fmt.Println("Start")
	var wg sync.WaitGroup
	// ticker := time.NewTicker(outputTime * time.Second)
	// _ = ticker
	superChan := make(chan Task, 10)
	doneChan := make(chan Task, 10)
	resultChan := make(chan ResultTasks)
	stop := make(chan bool)
	// Генерируем таски
	wg.Add(1)
	result := ResultTasks{DoneTask: []Task{}, ErrorTask: []Task{}}
	mu := sync.Mutex{}

	go tickerResult(&mu, &wg, &result, resultChan, stop)

	go taskGenerate(superChan, &wg)

	go func() {
		wg.Add(1)
		printResult(resultChan, &wg)
	}()

	go func() {
		for task_c := range superChan {
			go taskWorker(task_c, doneChan, &mu, &wg, &result) //!!!каналы сделать однонаправленные // go
			// fmt.Println(result.DoneTask)
		}
		// stop <- true
		fmt.Println("stop close")
		close(stop)
		close(doneChan)
		fmt.Println("close superchan")

	}()
	/*
		go func() {
			for r := range doneChan {
				// go func() {
				// 	for range ticker.C {
				// 		mu.Lock()
				// 		resultChan <- result
				// 		mu.Unlock()
				// 	}
				// }()
				fmt.Println("____2")
				// go func(r Task) {
				// fmt.Println(r)
				mu.Lock()
				result = append(result, r) // !!! rпереименновать
				// fmt.Println(result)
				mu.Unlock()

				wg.Done()
			}

			wg.Done()
			// wg.Done()
		}()*/

	wg.Wait()

	// time.Sleep(3 * time.Second)
	// !!! закрыть все не закрытые каналы
	finish := time.Now()
	fmt.Println("Finish - ", finish.Sub(start))

}
