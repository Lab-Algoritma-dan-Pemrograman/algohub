package objectstore

import (
	"errors"
	"testing"
	"time"
)

type fakeStore struct {
	enabled bool
	value   string
	err     error
	calls   int
}

func (f *fakeStore) Enabled() bool { return f.enabled }
func (f *fakeStore) Upload(string, []byte, string, bool) (string, error) {
	f.calls++
	return f.value, f.err
}
func (f *fakeStore) SignedUploadURL(string, time.Duration) (string, error) {
	f.calls++
	return f.value, f.err
}
func (f *fakeStore) SignedDownloadURL(string, time.Duration, string) (string, error) {
	f.calls++
	return f.value, f.err
}
func (f *fakeStore) Delete(string) error {
	f.calls++
	return f.err
}

func TestFailoverMemakaiPenyimpananBerikutnya(t *testing.T) {
	primary := &fakeStore{enabled: true, err: errors.New("upstream 402")}
	secondary := &fakeStore{enabled: true, value: "https://blob.test/file"}
	store := NewFailover(primary, secondary)

	got, err := store.Upload("file", []byte("x"), "text/plain", false)
	if err != nil || got != secondary.value {
		t.Fatalf("Upload() = %q, %v", got, err)
	}
	if primary.calls != 1 || secondary.calls != 1 {
		t.Fatalf("calls primary=%d secondary=%d", primary.calls, secondary.calls)
	}
}

func TestFailoverMelewatiClientNonaktif(t *testing.T) {
	disabled := &fakeStore{}
	active := &fakeStore{enabled: true, value: "ok"}
	store := NewFailover(disabled, active)

	got, err := store.SignedUploadURL("file", time.Minute)
	if err != nil || got != "ok" || disabled.calls != 0 || active.calls != 1 {
		t.Fatalf("hasil=%q err=%v calls=%d/%d", got, err, disabled.calls, active.calls)
	}
}
