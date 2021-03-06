// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy of
// the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

package kivik

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/internal/mock"
)

func TestRowsNext(t *testing.T) {
	tests := []struct {
		name     string
		rows     *Rows
		expected bool
	}{
		{
			name: "nothing more",
			rows: &Rows{
				iter: &iter{closed: true},
			},
			expected: false,
		},
		{
			name: "more",
			rows: &Rows{
				iter: &iter{
					feed: &mockIterator{
						NextFunc: func(_ interface{}) error { return nil },
					},
				},
			},
			expected: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.rows.Next()
			if result != test.expected {
				t.Errorf("Unexpected result: %v", result)
			}
		})
	}
}

func TestRowsErr(t *testing.T) {
	expected := "foo error"
	r := &Rows{
		iter: &iter{lasterr: errors.New(expected)},
	}
	err := r.Err()
	testy.Error(t, expected, err)
}

func TestRowsClose(t *testing.T) {
	expected := "close error"
	r := &Rows{
		iter: &iter{
			feed: &mockIterator{CloseFunc: func() error { return errors.New(expected) }},
		},
	}
	err := r.Close()
	testy.Error(t, expected, err)
}

func TestRowsIteratorNext(t *testing.T) {
	expected := "foo error"
	r := &rowsIterator{
		Rows: &mock.Rows{
			NextFunc: func(_ *driver.Row) error { return errors.New(expected) },
		},
	}
	var i driver.Row
	err := r.Next(&i)
	testy.Error(t, expected, err)
}

func TestRowsScanValue(t *testing.T) {
	tests := []struct {
		name     string
		rows     *Rows
		expected interface{}
		status   int
		err      string
	}{
		{
			name: "success",
			rows: &Rows{
				iter: &iter{
					ready: true,
					curVal: &driver.Row{
						ValueReader: strings.NewReader(`{"foo":123.4}`),
					},
				},
			},
			expected: map[string]interface{}{"foo": 123.4},
		},
		{
			name: "closed",
			rows: &Rows{
				iter: &iter{
					closed: true,
				},
			},
			status: http.StatusBadRequest,
			err:    "kivik: Iterator is closed",
		},
		{
			name: "row error",
			rows: &Rows{
				iter: &iter{
					ready: true,
					curVal: &driver.Row{
						Error: errors.New("row error"),
					},
				},
			},
			status: 500,
			err:    "row error",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var result interface{}
			err := test.rows.ScanValue(&result)
			testy.StatusError(t, test.err, test.status, err)
			if d := testy.DiffInterface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestRowsScanDoc(t *testing.T) {
	tests := []struct {
		name     string
		rows     *Rows
		expected interface{}
		status   int
		err      string
	}{
		{
			name: "old row",
			rows: &Rows{
				iter: &iter{
					ready: true,
					curVal: &driver.Row{
						Doc: []byte(`{"foo":123.4}`),
					},
				},
			},
			expected: map[string]interface{}{"foo": 123.4},
		},
		{
			name: "success",
			rows: &Rows{
				iter: &iter{
					ready: true,
					curVal: &driver.Row{
						DocReader: strings.NewReader(`{"foo":123.4}`),
					},
				},
			},
			expected: map[string]interface{}{"foo": 123.4},
		},
		{
			name: "closed",
			rows: &Rows{
				iter: &iter{
					closed: true,
				},
			},
			status: http.StatusBadRequest,
			err:    "kivik: Iterator is closed",
		},
		{
			name: "nil doc",
			rows: &Rows{
				iter: &iter{
					ready: true,
					curVal: &driver.Row{
						Doc: nil,
					},
				},
			},
			status: http.StatusBadRequest,
			err:    "kivik: doc is nil; does the query include docs?",
		},
		{
			name: "row error",
			rows: &Rows{
				iter: &iter{
					ready: true,
					curVal: &driver.Row{
						Error: errors.New("row error"),
					},
				},
			},
			status: 500,
			err:    "row error",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var result interface{}
			err := test.rows.ScanDoc(&result)
			testy.StatusError(t, test.err, test.status, err)
			if d := testy.DiffInterface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestRowsScanKey(t *testing.T) {
	tests := []struct {
		name     string
		rows     *Rows
		expected interface{}
		status   int
		err      string
	}{
		{
			name: "success",
			rows: &Rows{
				iter: &iter{
					ready: true,
					curVal: &driver.Row{
						Key: []byte(`{"foo":123.4}`),
					},
				},
			},
			expected: map[string]interface{}{"foo": 123.4},
		},
		{
			name: "closed",
			rows: &Rows{
				iter: &iter{
					closed: true,
				},
			},
			status: http.StatusBadRequest,
			err:    "kivik: Iterator is closed",
		},
		{
			name: "row error",
			rows: &Rows{
				iter: &iter{
					ready: true,
					curVal: &driver.Row{
						Error: errors.New("row error"),
					},
				},
			},
			status: 500,
			err:    "row error",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var result interface{}
			err := test.rows.ScanKey(&result)
			testy.StatusError(t, test.err, test.status, err)
			if d := testy.DiffInterface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestRowsGetters(t *testing.T) {
	id := "foo"
	key := []byte("[1234]")
	offset := int64(2)
	totalrows := int64(3)
	updateseq := "asdfasdf"
	r := &Rows{
		iter: &iter{
			ready: true,
			curVal: &driver.Row{
				ID:  id,
				Key: key,
			},
		},
		rowsi: &mock.Rows{
			OffsetFunc:    func() int64 { return offset },
			TotalRowsFunc: func() int64 { return totalrows },
			UpdateSeqFunc: func() string { return updateseq },
		},
	}

	t.Run("ID", func(t *testing.T) {
		result := r.ID()
		if id != result {
			t.Errorf("Unexpected result: %v", result)
		}
	})

	t.Run("Key", func(t *testing.T) {
		result := r.Key()
		if string(key) != result {
			t.Errorf("Unexpected result: %v", result)
		}
	})

	t.Run("Offset", func(t *testing.T) {
		result := r.Offset()
		if offset != result {
			t.Errorf("Unexpected result: %v", result)
		}
	})

	t.Run("TotalRows", func(t *testing.T) {
		result := r.TotalRows()
		if totalrows != result {
			t.Errorf("Unexpected result: %v", result)
		}
	})

	t.Run("UpdateSeq", func(t *testing.T) {
		result := r.UpdateSeq()
		if updateseq != result {
			t.Errorf("Unexpected result: %v", result)
		}
	})

	t.Run("Not Ready", func(t *testing.T) {
		r.ready = false

		t.Run("ID", func(t *testing.T) {
			result := r.ID()
			if result != "" {
				t.Errorf("Unexpected result: %v", result)
			}
		})

		t.Run("Key", func(t *testing.T) {
			result := r.Key()
			if result != "" {
				t.Errorf("Unexpected result: %v", result)
			}
		})
	})
}

func TestWarning(t *testing.T) {
	t.Run("Warner", func(t *testing.T) {
		expected := "test warning"
		r := newRows(context.Background(), &mock.RowsWarner{
			WarningFunc: func() string { return expected },
		})
		if w := r.Warning(); w != expected {
			t.Errorf("Warning\nExpected: %s\n  Actual: %s", expected, w)
		}
	})

	t.Run("NonWarner", func(t *testing.T) {
		r := newRows(context.Background(), &mock.Rows{})
		expected := ""
		if w := r.Warning(); w != expected {
			t.Errorf("Warning\nExpected: %s\n  Actual: %s", expected, w)
		}
	})
}

func TestQueryIndex(t *testing.T) {
	t.Run("QueryIndexer", func(t *testing.T) {
		expected := 100
		r := newRows(context.Background(), &mock.QueryIndexer{
			QueryIndexFunc: func() int { return expected },
		})
		if i := r.QueryIndex(); i != expected {
			t.Errorf("QueryIndex\nExpected %v\n  Actual: %v", expected, i)
		}
	})

	t.Run("Non QueryIndexer", func(t *testing.T) {
		r := newRows(context.Background(), &mock.Rows{})
		expected := 0
		if i := r.QueryIndex(); i != expected {
			t.Errorf("QueryIndex\nExpected: %v\n  Actual: %v", expected, i)
		}
	})
}

func TestBookmark(t *testing.T) {
	t.Run("Bookmarker", func(t *testing.T) {
		expected := "test bookmark"
		r := newRows(context.Background(), &mock.Bookmarker{
			BookmarkFunc: func() string { return expected },
		})
		if w := r.Bookmark(); w != expected {
			t.Errorf("Warning\nExpected: %s\n  Actual: %s", expected, w)
		}
	})
	t.Run("Non Bookmarker", func(t *testing.T) {
		r := newRows(context.Background(), &mock.Rows{})
		expected := ""
		if w := r.Bookmark(); w != expected {
			t.Errorf("Warning\nExpected: %s\n  Actual: %s", expected, w)
		}
	})
}
