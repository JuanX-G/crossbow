package queue

type itemData[T any] struct {
	item   T
	dropCh chan error
}

func WrapInItemData[T any](item T, dropCh chan error) itemData[T] {
	return itemData[T]{item: item, dropCh: dropCh}
}
