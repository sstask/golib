package stnet

type BufferPool struct {
	bufSize  int // size of each buffer
	freeList chan []byte
}

const maxNBuf = 4096

var bufferPool = NewBufferPool(maxNBuf, MsgBuffSize)

func NewBufferPool(n, bufsize int) *BufferPool {
	return &BufferPool{
		bufSize:  bufsize,
		freeList: make(chan []byte, n),
	}
}

func (bp *BufferPool) Alloc(bufsize int) (b []byte) {
	if bufsize > bp.bufSize {
		b = make([]byte, bufsize)
	} else {
		select {
		case b = <-bp.freeList:
		default:
			b = make([]byte, bp.bufSize)
		}
	}
	return
}

func (bp *BufferPool) Free(b []byte) {
	if len(b) != bp.bufSize {
		return
	}
	select {
	case bp.freeList <- b:
	default:
	}
	return
}
