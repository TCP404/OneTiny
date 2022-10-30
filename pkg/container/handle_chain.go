package container

type Handler interface {
	Handle() error
}

type handleNode struct {
	Handler
	Next *handleNode
}

type Chain interface {
	AddToHead(Handler) Chain
	AddToTail(Handler) Chain
	Iterator() error
}

var _ Chain = (*handleChain)(nil)

func NewHandleChain() Chain {
	return &handleChain{}
}

type handleChain struct {
	head *handleNode
	tail *handleNode
}

func (c *handleChain) AddToHead(h Handler) Chain {
	node := &handleNode{
		Handler: h,
		Next:    c.head,
	}
	if c.head == nil {
		c.head, c.tail = node, node
		return c
	}
	c.head = node
	return c
}

func (c *handleChain) AddToTail(h Handler) Chain {
	node := &handleNode{
		Handler: h,
		Next:    nil,
	}
	if c.head == nil || c.tail == nil {
		c.head, c.tail = node, node
		return c
	}
	c.tail.Next = node
	return c
}

func (c *handleChain) Iterator() error {
	curr := c.head
	for curr != nil {
		if err := curr.Handle(); err != nil {
			return err
		}
		curr = curr.Next
	}
	return nil
}
