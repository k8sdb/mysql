// Copyright 2013 Xing Xing <mikespook@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a commercial
// license that can be found in the LICENSE file.

package idgen

import (
    "testing"
)

func TestObjectId(t *testing.T) {
	id := NewObjectId()
	a := id.Id(); b := id.Id()
	if a == b {
		t.Errorf("%s is equal to %s", a, b)
	}
}

func TestAutoIncId(t *testing.T) {
	id := NewAutoIncId()
	a := id.Id(); b := id.Id()
	if a.(int) + 1 != b {
		t.Errorf("%d's next is not %d", a, b)
	}
}
