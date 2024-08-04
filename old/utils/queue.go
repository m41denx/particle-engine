package utils

type any *interface{}

type Queue struct {
	queue []any
}

func NewQueue() *Queue {
	return &Queue{}
}

func (q *Queue) Add(e any) {
	q.queue = append(q.queue, e)
}

func (q *Queue) Pop() any {
	if q.Size() == 0 {
		return nil
	}
	e := q.queue[0]
	q.queue = q.queue[1:]
	return e
}

func (q *Queue) Size() int {
	return len(q.queue)
}

type Stack struct {
	queue []any
}

func NewStack() *Stack {
	return &Stack{}
}

func (s *Stack) Add(e any) {
	q := []any{e}
	s.queue = append(q, s.queue...)
}

func (s *Stack) Pop() any {
	if s.Size() == 0 {
		return nil
	}
	e := s.queue[0]
	s.queue = s.queue[1:]
	return e
}

func (s *Stack) Size() int {
	return len(s.queue)
}
