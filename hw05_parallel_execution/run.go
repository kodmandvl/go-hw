package hw05parallelexecution

import (
	"errors"
	"sync"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
func Run(tasks []Task, n, m int) error {
	if m <= 0 {
		// Значение m <= 0 трактуется на усмотрение программиста. Выбираем вариант:
		// считать это как "максимум 0 ошибок", значит функция всегда будет возвращать ErrErrorsLimitExceeded;
		return ErrErrorsLimitExceeded
	}

	wg := sync.WaitGroup{}  // WaitGroup для n горутин и ожидания их завершения
	lockCnt := sync.Mutex{} // Mutex для защиты счётчика ошибок
	lockInd := sync.Mutex{} // Mutex для защиты индекса выбираемых тасок
	var cnt int             // счётчик ошибок
	var ind int             // глобальный безопасный индекс для выполняемых тасок, выбираемых из слайса
	tLen := len(tasks)      // размер слайса тасков (чтобы не вычислять его каждый раз, когда он понадобится)

	// Цикл запуска n горутин:
	for i := 0; i < n; i++ {
		wg.Add(1) // Добавляем в WaitGroup
		go func() {
			defer wg.Done() // Через defer в конце выполнения i-ой горутины уменьшаем счётчик WaitGroup
			// Получаем таски из слайса тасков:
			for {
				// Манипуляции с индексом слайса проводим под защитой (через mutex):
				lockInd.Lock()
				i := ind
				ind++
				lockInd.Unlock()
				if i >= tLen {
					return
				}
				// Манипуляции со счётчиком ошибок проводим под защитой (через mutex):
				lockCnt.Lock()
				if cnt >= m {
					lockCnt.Unlock()
					return
				}
				lockCnt.Unlock()
				e := tasks[i]()
				lockCnt.Lock()
				if e != nil {
					cnt++
				}
				lockCnt.Unlock()
			}
		}()
	}

	wg.Wait() // Ожидание окончания выполнения n горутин

	// Если было превышено m, возвращаем ошибку ErrErrorsLimitExceeded:
	if cnt >= m {
		return ErrErrorsLimitExceeded
	}

	// Если ранее не вышли из функции, вернув ErrErrorsLimitExceeded, то всё хорошо, возвращаем nil:
	return nil
}
