package queue

import (
	"testing"
	"time"
)

func TestPushPop(t *testing.T) {
	q := NewQueue(2, 4, BlockPolicy[int]{})
	q.Push(1)
	q.Push(2)

	v, ok := q.Pop()
	if !ok {
		t.Fatalf("pop returned false but value inside were expected. Value returned: %d", v)
	}
	if v != 1 {
		t.Fatalf("pop returned wrong value. expected: 1, found: %d", v)
	}
}

func TestBlockPolicy(t *testing.T) {
	q := NewQueue(2, 2, BlockPolicy[int]{})
	q.Push(1)
	q.Push(2)
	doneCh := make(chan struct{})
	pushFn := func () {
		q.Push(3)
		doneCh <- struct {}{}
	}
	go pushFn()

	select {
	case <-doneCh:
		t.Fatalf("Should have been blocked")
	default:
	}

	time.Sleep(time.Millisecond * 150)
	q.Pop()
	time.Sleep(time.Millisecond * 25)
	select {
	case <-doneCh:
	default:
		t.Fatalf("Should have unblocked now")
	}

}

type policyTest struct {
	input []int
	endState []int
}

func TestPolicies(t *testing.T) {
	tests := map[MailboxPolicy[int]]policyTest{
		DropNewestPolicy[int]{}: policyTest{input: []int{1, 2, 3}, endState: []int{1, 2}},
		DropOldestPolicy[int]{}: policyTest{input: []int{1, 2, 3}, endState: []int{2, 3}},
	}
	for policy, test := range tests {
		q := NewQueue(2, 2, policy)
		for _, v := range test.input {
			q.Push(v)
		}
		for _, vs := range test.endState {
			if vp, _ := q.Pop(); vp != vs {
				t.Fatalf("Queue with policy %#v returned a wrong value; expcted: %d, found: %d", policy, vs, vp)
			}
		}
	}
}
