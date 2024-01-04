package game

// actorQueue implements a priority queue for actors based on next scheduled AI
// action.
type actorQueue []*Actor

func (q actorQueue) Len() int { return len(q) }

func (q actorQueue) Less(i, j int) bool {
	return q[i].NextThink.Before(q[j].NextThink)
}

func (q actorQueue) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
	q[i].pqIdx = i
	q[j].pqIdx = j
}

func (q *actorQueue) Push(x any) {
	n := len(*q)
	item := x.(*Actor)
	item.pqIdx = n
	*q = append(*q, item)
}

func (q *actorQueue) Pop() any {
	old := *q
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.pqIdx = -1
	*q = old[0 : n-1]
	return item
}
