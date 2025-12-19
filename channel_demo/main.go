package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func main() {
	// Инициализируем генератор случайных чисел
	rand.Seed(time.Now().UnixNano())
	
	// Создаем два канала для передачи целых чисел
	channel1 := make(chan int, 10) // буферизованный канал для предотвращения блокировки
	channel2 := make(chan int, 10) // буферизованный канал для предотвращения блокировки

	// Счетчики для отслеживания количества полученных сообщений из каждого канала
	count1 := 0
	count2 := 0

	// Мьютекс для безопасного доступа к счетчикам из разных горутин
	var mutex sync.Mutex

	// Запускаем горутины для отправки данных в каналы
	go func() {
		for i := 1; i <= 100; i++ {
			select {
			case channel1 <- i:
				// Успешно отправили данные в канал 1
			default:
				// Канал заблокирован, но мы не ждем, просто продолжаем
			}
			time.Sleep(time.Millisecond * 5) // быстрее заполняем канал 1
		}
	}()

	go func() {
		for i := 1; i <= 100; i++ {
			select {
			case channel2 <- i:
				// Успешно отправили данные в канал 2
			default:
				// Канал заблокирован, но мы не ждем, просто продолжаем
			}
			time.Sleep(time.Millisecond * 10) // медленнее заполняем канал 2
		}
	}()

	// Функция для получения данных из каналов с 2:1 соотношением
	// Используем стратегию: когда оба канала доступны, с вероятностью 2/3 читаем из канала 1, 1/3 - из канала 2
	receiveWithRatio := func() {
		for i := 0; i < 500; i++ { // увеличиваем количество итераций для лучшей статистики
			// Используем неблокирующий select с вероятностным выбором
			select {
			case val1, ok1 := <-channel1:
				if ok1 {
					// Если доступен только канал 1
					mutex.Lock()
					count1++
					fmt.Printf("Получено из канала 1: %d (Всего: %d)\n", val1, count1)
					mutex.Unlock()
				}
			case val2, ok2 := <-channel2:
				if ok2 {
					// Если доступен только канал 2
					mutex.Lock()
					count2++
					fmt.Printf("Получено из канала 2: %d (Всего: %d)\n", val2, count2)
					mutex.Unlock()
				}
			default:
				// Ни один из каналов не доступен, используем вероятностный выбор при возможности
				// Проверим снова с небольшой задержкой
				time.Sleep(time.Millisecond * 1)
				
				// Попробуем вероятностный выбор
				select {
				case val1, ok1 := <-channel1:
					if ok1 {
						mutex.Lock()
						count1++
						fmt.Printf("Получено из канала 1: %d (Всего: %d)\n", val1, count1)
						mutex.Unlock()
					}
				default:
					// Канал 1 не доступен, пробуем канал 2
					select {
					case val2, ok2 := <-channel2:
						if ok2 {
							mutex.Lock()
							count2++
							fmt.Printf("Получено из канала 2: %d (Всего: %d)\n", val2, count2)
							mutex.Unlock()
						}
					default:
						// Оба канала недоступны, пробуем оба с вероятностью 2:1
						// С вероятностью 2/3 пробуем канал 1 первым
						if rand.Intn(3) < 2 { // 2 из 3 случаев (~66%) - канал 1
							select {
							case val1, ok1 := <-channel1:
								if ok1 {
									mutex.Lock()
									count1++
									fmt.Printf("Получено из канала 1: %d (Всего: %d)\n", val1, count1)
									mutex.Unlock()
								}
							default:
								// Канал 1 недоступен, пробуем канал 2
								select {
								case val2, ok2 := <-channel2:
									if ok2 {
										mutex.Lock()
										count2++
										fmt.Printf("Получено из канала 2: %d (Всего: %d)\n", val2, count2)
										mutex.Unlock()
									}
								default:
									// Оба канала недоступны
								}
							}
						} else { // 1 из 3 случаев (~33%) - канал 2
							select {
							case val2, ok2 := <-channel2:
								if ok2 {
									mutex.Lock()
									count2++
									fmt.Printf("Получено из канала 2: %d (Всего: %d)\n", val2, count2)
									mutex.Unlock()
								}
							default:
								// Канал 2 недоступен, пробуем канал 1
								select {
								case val1, ok1 := <-channel1:
									if ok1 {
										mutex.Lock()
										count1++
										fmt.Printf("Получено из канала 1: %d (Всего: %d)\n", val1, count1)
										mutex.Unlock()
									}
								default:
									// Оба канала недоступны
								}
							}
						}
					}
				}
			}
		}
	}

	// Запускаем функцию получения данных
	receiveWithRatio()

	// Ждем немного, чтобы убедиться, что все данные обработаны
	time.Sleep(time.Second * 1)

	// Выводим финальную статистику
	mutex.Lock()
	defer mutex.Unlock()
	fmt.Printf("\nФинальная статистика:\n")
	fmt.Printf("Соотношение каналов: %d:%d\n", count1, count2)
	if count2 != 0 {
		ratio := float64(count1) / float64(count2)
		fmt.Printf("Отношение каналов 1 к 2: %.2f:1\n", ratio)
	} else {
		fmt.Println("Канал 2 не использовался")
	}
}