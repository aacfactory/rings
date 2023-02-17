/*
 * Copyright 2021 Wang Min Xiang
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * 	http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package rings

import (
	"bytes"
	"sync"
)

type node[E Entry] struct {
	next, prev *node[E]
	value      E
}

func New[E Entry](key string, values ...E) (r *Ring[E]) {
	r = &Ring[E]{
		mutex: sync.RWMutex{},
		key:   key,
		head:  nil,
		size:  0,
	}
	if values != nil && len(values) > 0 {
		for _, value := range values {
			r.Push(value)
		}
	}
	return
}

type Ring[E Entry] struct {
	mutex sync.RWMutex
	key   string
	head  *node[E]
	size  int
}

func (r *Ring[E]) Key() (key string) {
	key = r.key
	return
}

func (r *Ring[E]) Push(v E) (ok bool) {
	r.mutex.Lock()
	_, has := r.get(v.Key())
	if has {
		r.mutex.Unlock()
		return
	}
	e := &node[E]{
		value: v,
	}
	if r.head == nil {
		e.next = e
		e.prev = e
		r.head = e
	} else {
		prev := r.head.prev
		prev.next = e
		e.prev = prev
		e.next = r.head
		r.head.prev = e
	}
	r.size++
	ok = true
	r.mutex.Unlock()
	return
}

func (r *Ring[E]) Pop() (e E, ok bool) {
	r.mutex.Lock()
	if r.head == nil {
		r.mutex.Unlock()
		return
	}
	head := r.head
	e = head.value
	r.size--
	if r.size == 0 {
		head = nil
	} else {
		next := head.next
		next.prev = head.prev
		head = next
	}
	r.head = head
	r.mutex.Unlock()
	return
}

func (r *Ring[E]) Remove(key string) {
	r.mutex.Lock()
	if r.head == nil {
		r.mutex.Unlock()
		return
	}
	head := r.head
	for i := 0; i < r.size; i++ {
		e := r.next()
		if e.value.Key() == key {
			if e.prev.value.Key() == key && e.next.value.Key() == key {
				head = nil
				break
			}
			prev := e.prev
			next := e.next
			prev.next = next
			if head.value.Key() == key {
				head = next
			}
			break
		}
	}
	r.head = head
	r.size--
	r.mutex.Unlock()
}

func (r *Ring[E]) Next() (value E) {
	r.mutex.RLock()
	if r.size == 0 {
		r.mutex.RUnlock()
		return
	}
	value = r.next().value
	r.mutex.RUnlock()
	return
}

func (r *Ring[E]) SeekTo(key string) (has bool) {
	r.mutex.RLock()
	if r.size == 0 {
		r.mutex.RUnlock()
		return
	}
	for i := 0; i < r.size; i++ {
		n := r.next()
		if n.value.Key() == key {
			r.head = n
			has = true
			break
		}
	}
	r.mutex.RUnlock()
	return
}

func (r *Ring[E]) Head() (value E, has bool) {
	r.mutex.RLock()
	if r.size == 0 {
		r.mutex.RUnlock()
		return
	}
	value = r.head.value
	has = true
	r.mutex.RUnlock()
	return
}

func (r *Ring[E]) Get(key string) (value E, has bool) {
	r.mutex.RLock()
	if r.size == 0 {
		r.mutex.RUnlock()
		return
	}
	var n *node[E] = nil
	n, has = r.get(key)
	if has {
		value = n.value
	}
	r.mutex.RUnlock()
	return
}

func (r *Ring[E]) get(key string) (value *node[E], has bool) {
	if r.size == 0 {
		return
	}
	head := r.head
	for i := 0; i < r.size; i++ {
		n := r.next()
		if n.value.Key() == key {
			value = n
			has = true
			break
		}
	}
	r.head = head
	return
}

func (r *Ring[E]) Len() (size int) {
	r.mutex.RLock()
	size = r.size
	r.mutex.RUnlock()
	return
}

func (r *Ring[E]) String() (value string) {
	r.mutex.RLock()
	p := bytes.NewBufferString("")
	_ = p.WriteByte('[')
	for i := 0; i < r.size; i++ {
		e := r.next()
		if i == 0 {
			_, _ = p.WriteString(e.value.Key())
		} else {
			_, _ = p.WriteString(", ")
			_, _ = p.WriteString(e.value.Key())
		}
	}
	_ = p.WriteByte(']')
	value = p.String()
	r.mutex.RUnlock()
	return
}

func (r *Ring[E]) next() (node *node[E]) {
	node = r.head
	r.head = r.head.next
	return
}
