// Copyright 2019 The OpenSDS Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package opensds

import (
	"fmt"
	"reflect"
	"testing"
)

var assertTestResult = func(t *testing.T, got, expected interface{}) {
	t.Helper()
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("expected %v, got %v\n", expected, got)
	}
}

type Obj struct {
	Id     string
	Name   string
	Status string
}

type ObjWithNoStatus struct {
	Id string
}

func TestWaitForStatusStable(t *testing.T) {
	objList := []*Obj{
		&Obj{
			Id:     "1",
			Name:   "for test",
			Status: "available",
		},
		&Obj{
			Id:     "2",
			Name:   "for test",
			Status: "error",
		},
		&Obj{
			Id:     "4",
			Name:   "for test",
			Status: "binding",
		},
	}

	f := func(id string) (interface{}, error) {
		for _, v := range objList {
			if v.Id == id {
				return v, nil
			}
		}
		return nil, fmt.Errorf("cannot find %s", id)
	}

	t.Run("Object status is available", func(t *testing.T) {
		result, _ := waitForStatusStable("1", f)
		assertTestResult(t, objList[0], result)
	})

	t.Run("Object status is error", func(t *testing.T) {
		_, err := waitForStatusStable("2", f)
		assertTestResult(t, "status is error but not available", err.Error())
	})

	t.Run("Object does not exist", func(t *testing.T) {
		_, err := waitForStatusStable("3", f)
		assertTestResult(t, "cannot find 3", err.Error())
	})

	t.Run("Object status is invalid", func(t *testing.T) {
		_, err := waitForStatusStable("4", f)
		assertTestResult(t, "invalid status: binding", err.Error())
	})

	t.Run("Object struct has no status filed", func(t *testing.T) {
		o := &ObjWithNoStatus{Id: "1"}
		f := func(id string) (interface{}, error) {
			return o, nil
		}

		_, err := waitForStatusStable("1", f)
		assertTestResult(t, fmt.Sprintf("cannot find status from struct %v", o), err.Error())
	})

	t.Run("input is empty string", func(t *testing.T) {
		f := func(id string) (interface{}, error) {
			return "", nil
		}

		_, err := waitForStatusStable("1", f)
		assertTestResult(t, "input must be struct", err.Error())
	})

	t.Run("input is nil", func(t *testing.T) {
		f := func(id string) (interface{}, error) {
			var o *Obj
			return o, nil
		}

		_, err := waitForStatusStable("1", f)
		assertTestResult(t, "input cannot be nil", err.Error())
	})
}
