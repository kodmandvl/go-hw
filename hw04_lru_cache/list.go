package hw04lrucache

type List interface {
	Len() int
	Front() *ListItem
	Back() *ListItem
	PushFront(v interface{}) *ListItem
	PushBack(v interface{}) *ListItem
	Remove(i *ListItem)
	MoveToFront(i *ListItem)
}

type ListItem struct {
	Value interface{}
	Next  *ListItem
	Prev  *ListItem
}

type list struct {
	length    int
	listBegin *ListItem
	listEnd   *ListItem
}

func NewList() List {
	return new(list)
}

// Len() int // длина списка.
func (l *list) Len() int {
	return l.length
}

// Front() *ListItem // первый элемент списка.
func (l *list) Front() *ListItem {
	return l.listBegin
}

// Back() *ListItem // последний элемент списка.
func (l *list) Back() *ListItem {
	return l.listEnd
}

// PushFront(v interface{}) *ListItem // добавить значение в начало.
func (l *list) PushFront(v interface{}) *ListItem {
	// Новый элемент списка:
	newElem := &ListItem{
		Value: v,
		Next:  l.listBegin,
		Prev:  nil,
	}

	// Если l.listBegin != nil (т.е. если по сути список был не пуст),
	// то теперь у бывшего "начала списка" предыдущий элемент теперь - newElem:
	if l.listBegin != nil {
		l.listBegin.Prev = newElem
	} else {
		// Иначе (т.е. если спиок был пуст) теперь это будет последний элемент:
		l.listEnd = newElem
	}

	// При этом добавляемый элемент также будет первым в списке:
	l.listBegin = newElem

	// Длина списка увеличивается на 1:
	l.length++

	// Теперь после всех необходимых проверок и действий возвращаем этот новый элемент:
	return newElem
}

// PushBack(v interface{}) *ListItem // добавить значение в конец.
func (l *list) PushBack(v interface{}) *ListItem {
	// Новый элемент списка:
	newElem := &ListItem{
		Value: v,
		Next:  nil,
		Prev:  l.listEnd,
	}

	// Если l.listEnd != nil (т.е. если по сути список был не пуст),
	// то теперь у бывшего "конца списка" следующий элемент - newElem:
	if l.listEnd != nil {
		l.listEnd.Next = newElem
	} else {
		// Иначе (т.е. если спиок был пуст) теперь это будет первый элемент:
		l.listBegin = newElem
	}

	// При этом добавляемый элемент также будет последним в списке:
	l.listEnd = newElem

	// Длина списка увеличивается на 1:
	l.length++

	// Теперь после всех необходимых проверок и действий возвращаем этот новый элемент:
	return newElem
}

// Remove(i *ListItem) // удалить элемент.
func (l *list) Remove(i *ListItem) {
	// Если предыдущего элемента нет, значит этот был первым в списке:
	if i.Prev == nil {
		// Тогда теперь следующий элемент будет первым:
		l.listBegin = i.Next
		// и у него не будет предыдущего элемента:
		if l.listBegin != nil {
			l.listBegin.Prev = nil
		}
	} else {
		// Иначе следующий у предыдущего - это следующий у удаляемого
		// (убрали удаляемый элемент из цепочки):
		i.Prev.Next = i.Next
	}

	// Если следующего элемента нет, значит этот был последним в списке:
	if i.Next == nil {
		// Тогда теперь предыдущий элемент будет последним:
		l.listEnd = i.Prev
		// и у него не будет следующего элемента:
		if l.listEnd != nil {
			l.listEnd.Next = nil
		}
	} else {
		// Иначе предыдущий у следующего - это предыдущий у удаляемого
		// (убрали удаляемый элемент из цепочки):
		i.Next.Prev = i.Prev
	}

	// Не забываем уменьшить длину списка:
	l.length--
}

// MoveToFront(i *ListItem) // переместить элемент в начало.
func (l *list) MoveToFront(i *ListItem) {
	// Если этот элемент - и так начало списка, ничего делать не нужно:
	if i == l.listBegin {
		return
	}

	// Убираем перемещаемый элемент из цепочки:
	i.Prev.Next = i.Next
	if i.Next != nil {
		i.Next.Prev = i.Prev
	} else {
		l.listEnd = i.Prev
	}

	// Размещаем его в начале и связываем с бывшим начальным элементом:
	l.listBegin.Prev = i
	i.Prev = nil
	i.Next = l.listBegin
	l.listBegin = i
}
