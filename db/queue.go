package db

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"gopkg.in/rroy233/logger.v2"
	"sync"
	"time"
)

var (
	ErrorQueueFull  = errors.New("ErrorQueueFull")
	ErrorQueueEmpty = errors.New("ErrorQueueEmpty")
	ErrorNotAllowed = errors.New("ErrorNotAllowed")
	ErrorNotFound   = errors.New("ErrorNotFound")
	ErrorAborted    = errors.New("ErrorAborted")
)

// QStruct 队列规则：
//
// 1.凭借用户UID入队，单个用户的不同请求可同时存在于队列中，返回QItem作为凭证。
//
// 2.在队列过程中可调用Abort()进行弃权，则队列的后续用户将被允许跳过自己进行出队。
//
// 3.出队时调用QItem的Dequeue()方法，若队首不是自己或前方存在未弃权的用户，则出队失败返回ErrorNotAllowed。
//
// 4.若业务过程结束后希望立即完成出队，可先调用QItem的Abort()方法，进行弃权登记，则可立即允许完成出队。
type QStruct struct {
	data []*QItem
	size int
	head int
	tail int
	lock sync.Mutex
}

type QItem struct {
	UUID       string
	queueIndex int
	uid        int64
	addTime    int64
	abort      bool
}

var queue *QStruct
var maxQueueSize int
var QueueTimeout int64

const queueCleanerInterval = 10 * time.Second

func initQueue(maxSize int) {
	if maxSize == 0 {
		maxQueueSize = 5
	} else {
		maxQueueSize = maxSize
	}

	//队列任务超时时间
	QueueTimeout = 30

	queue = new(QStruct)
	queue.size = 0
	queue.head = 0
	queue.tail = 0
	queue.data = make([]*QItem, maxQueueSize)
	go queueCleaner()
	return
}

// 定期清除队列中超时项
func queueCleaner() {
	for true {
		queue.lock.Lock()
		for i := 0; i < maxQueueSize; i++ {
			if queue.data[i] == nil {
				continue
			}
			if time.Now().Unix()-queue.data[i].addTime > QueueTimeout {
				queue.data[i].abort = true
			}
		}
		queue.lock.Unlock()

		p := queue.head
		for queue.data[p] != nil && queue.data[p].abort == true {
			queue.pop()
			p = (p + 1) % maxQueueSize
		}

		time.Sleep(queueCleanerInterval)
	}
}

// EnQueue 入队
//
// 返回*QItem ，可随时调用Abort()进行弃权操作
//
// 若队伍已满则返回ErrorQueueFull
func EnQueue(UID int64) (*QItem, error) {
	if queue.size == maxQueueSize {
		return &QItem{}, ErrorQueueFull
	}
	item := &QItem{
		UUID:    uuid.New().String(),
		uid:     UID,
		addTime: time.Now().Unix(),
		abort:   false,
	}
	queue.push(item)
	return item, nil
}

// DeQueue 出队表示已完成等待
func (q *QItem) DeQueue() error {
	if queue.size == 0 {
		return ErrorQueueEmpty
	}
	if queue.data[queue.head].UUID != q.UUID && queue.data[queue.head].abort != true {
		return ErrorNotAllowed
	}
	var popItem *QItem
	success := false
	for true {
		if queue.size == 0 {
			break
		}
		if queue.data[queue.head].UUID == q.UUID || queue.data[queue.head].abort == true {
			//队首是自己 或 队首已弃权，则弹出
			popItem = queue.pop()
			if popItem.UUID == q.UUID {
				success = true
			}
			continue
		}
		break
	}
	if success == false {
		return ErrorNotFound
	}
	return nil
}

// FindQueueItemByUUID 通过UUID找回QItem
//
// 若不在队伍中则返回ErrorNotFound
//
// 若已弃权返回ErrorAborted
func FindQueueItemByUUID(UUID string) (*QItem, error) {
	return queue.find(UUID)
}

// QueryFront 查询前面的用户数
//
// 返回-1表示不存在或已弃权
func (q *QItem) QueryFront() int {
	front, exist := queue.findRelIndex(q.UUID)
	if exist == false {
		return -1
	}
	return front
}

// Abort 弃权
func (q *QItem) Abort() {
	q.abort = true
	return
}

// IsAbort 查询是否已弃权
func (q *QItem) IsAbort() bool {
	return q.abort
}

func (q *QStruct) push(item *QItem) {
	q.lock.Lock()
	defer q.lock.Unlock()
	item.queueIndex = q.tail
	q.data[q.tail] = item
	q.size++
	q.tail = (q.tail + 1) % maxQueueSize
	return
}

func (q *QStruct) pop() *QItem {
	q.lock.Lock()
	defer q.lock.Unlock()
	item := q.data[q.head]
	q.data[q.head] = nil
	q.size--
	q.head = (q.head + 1) % maxQueueSize
	if q.size == 0 {
		q.head = 0
		q.tail = 0
	}
	return item
}

func (q *QStruct) debugPrint() {
	q.lock.Lock()
	defer q.lock.Unlock()
	text := ""
	text += fmt.Sprintf("[Size=%d,head=%d,tail=%d]->[\n", q.size, q.head, q.tail)
	for i := 0; i < maxQueueSize; i++ {
		text += fmt.Sprintf("\t")
		if q.data[i] == nil {
			text += fmt.Sprintf("<nil>")
		} else {
			text += fmt.Sprintf("UUID=%s\tUID=%d\tabort=%v\tadd_time=%d\tqueueIndex=%d",
				q.data[i].UUID, q.data[i].uid, q.data[i].abort, q.data[i].addTime, q.data[i].queueIndex,
			)
		}
		text += "\n"
	}
	if logger.Debug == nil {
		logger.New(&logger.Config{StdOutput: true})
	}
	logger.Debug.Println(text)
	return
}

// 找相对于队头的索引
// int为前面的个数
// bool为是否找到
func (q *QStruct) findRelIndex(UUID string) (int, bool) {
	q.lock.Lock()
	defer q.lock.Unlock()
	if q.size == 0 {
		return -1, false
	}
	i := 0
	abortNum := 0
	found := false
	if q.data[q.head] != nil {
		if q.data[q.head].UUID == UUID && q.data[q.head].abort != true {
			return 0, true
		}
	}
	if q.data[q.head] != nil && q.data[q.head].abort == true {
		abortNum++
	}
	for i = (q.head + 1) % maxQueueSize; i != q.tail; i = (i + 1) % maxQueueSize {
		if queue.data[i] == nil {
			continue
		}
		if q.data[i].abort == true {
			abortNum++
		}
		if q.data[i].UUID == UUID {
			found = true
			break
		}
	}
	//未找到
	if found == false {
		return -1, false
	}
	//自身为弃权的
	if q.data[i].abort == true {
		return -1, true
	}

	i = (i - q.head + maxQueueSize) % maxQueueSize
	return i - abortNum, true
}

func (q *QStruct) find(UUID string) (*QItem, error) {
	q.lock.Lock()
	defer q.lock.Unlock()
	if q.size == 0 {
		return nil, ErrorQueueEmpty
	}
	if q.data[q.head] != nil && q.data[q.head].UUID == UUID {
		if q.data[q.head].abort != true {
			return q.data[q.head], nil
		} else {
			return nil, ErrorAborted
		}
	}
	i := 0
	found := false
	for i = (q.head + 1) % maxQueueSize; i != q.tail; i = (i + 1) % maxQueueSize {
		if queue.data[i] == nil {
			continue
		}
		if q.data[i].UUID == UUID {
			found = true
			break
		}
	}
	//未找到
	if found == false {
		return nil, ErrorNotFound
	}
	//自身为弃权的
	if q.data[i].abort == true {
		return nil, ErrorAborted
	}
	return q.data[i], nil
}
