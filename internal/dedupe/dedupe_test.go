package dedupe

import (
	"context"
	"errors"
	"testing"

	"github.com/xoxyn/semul/internal/domain"
)

// fakeStore is an in-memory Store.
type fakeStore struct {
	companies map[string]bool
	emails    map[string]bool
	checkErr  error
	markErr   error
}

func newFake() *fakeStore {
	return &fakeStore{companies: map[string]bool{}, emails: map[string]bool{}}
}

func (f *fakeStore) Contacted(_ context.Context, p domain.Prospect) (bool, error) {
	if f.checkErr != nil {
		return false, f.checkErr
	}
	return f.companies[p.CompanyKey()] || f.emails[p.EmailKey()], nil
}

func (f *fakeStore) MarkSent(_ context.Context, job domain.EmailJob) error {
	if f.markErr != nil {
		return f.markErr
	}
	f.companies[job.CompanyKey()] = true
	f.emails[job.EmailKey()] = true
	return nil
}

func prospect(company, email string) domain.Prospect {
	return domain.Prospect{Company: company, Email: email}
}

func TestFilterSkipsPreviouslyContacted(t *testing.T) {
	store := newFake()
	store.companies["acme logistics"] = true

	fresh, skipped, err := Filter(context.Background(), store, []domain.Prospect{
		prospect("ACME  Logistics", "new@acme.test"), // same company, different case/spacing
		prospect("Globex", "g@globex.test"),
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(fresh) != 1 || fresh[0].Company != "Globex" {
		t.Errorf("fresh = %v, want only Globex", fresh)
	}
	if len(skipped) != 1 {
		t.Errorf("skipped = %v, want 1", skipped)
	}
}

func TestFilterSkipsPreviouslyContactedEmail(t *testing.T) {
	store := newFake()
	store.emails["hello@initech.test"] = true

	fresh, skipped, _ := Filter(context.Background(), store, []domain.Prospect{
		prospect("Initech Renamed", " Hello@Initech.test "),
	})

	if len(fresh) != 0 || len(skipped) != 1 {
		t.Errorf("fresh=%v skipped=%v, want all skipped", fresh, skipped)
	}
}

func TestFilterSkipsDuplicatesWithinSameFile(t *testing.T) {
	fresh, skipped, err := Filter(context.Background(), newFake(), []domain.Prospect{
		prospect("Acme", "a@acme.test"),
		prospect("acme", "other@acme.test"), // same company
		prospect("Globex", "a@acme.test"),   // same email
		prospect("Initech", "i@initech.test"),
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(fresh) != 2 || fresh[0].Company != "Acme" || fresh[1].Company != "Initech" {
		t.Errorf("fresh = %v, want Acme and Initech", fresh)
	}
	if len(skipped) != 2 {
		t.Errorf("skipped = %d, want 2", len(skipped))
	}
}

func TestFilterStoreError(t *testing.T) {
	store := newFake()
	store.checkErr = errors.New("db down")

	_, _, err := Filter(context.Background(), store, []domain.Prospect{prospect("Acme", "a@acme.test")})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestWrapSendRecordsSuccess(t *testing.T) {
	store := newFake()
	calls := 0
	send := WrapSend(store, func(context.Context, domain.EmailJob) error {
		calls++
		return nil
	})
	job := domain.EmailJob{Prospect: prospect("Acme", "a@acme.test")}

	if err := send(context.Background(), job); err != nil {
		t.Fatal(err)
	}

	// Second attempt must be refused without calling the real sender.
	err := send(context.Background(), job)
	if !errors.Is(err, domain.ErrAlreadyContacted) {
		t.Fatalf("err = %v, want ErrAlreadyContacted", err)
	}
	if calls != 1 {
		t.Errorf("real sender called %d times, want 1", calls)
	}
}

func TestWrapSendDoesNotRecordFailures(t *testing.T) {
	store := newFake()
	fail := true
	send := WrapSend(store, func(context.Context, domain.EmailJob) error {
		if fail {
			return errors.New("smtp down")
		}
		return nil
	})
	job := domain.EmailJob{Prospect: prospect("Acme", "a@acme.test")}

	if err := send(context.Background(), job); err == nil {
		t.Fatal("expected smtp error")
	}

	// A failed send is not history, so retrying must be allowed.
	fail = false
	if err := send(context.Background(), job); err != nil {
		t.Fatalf("retry failed: %v", err)
	}
}

func TestWrapSendFailsClosedWhenHistoryUnavailable(t *testing.T) {
	store := newFake()
	store.checkErr = errors.New("db down")
	called := false
	send := WrapSend(store, func(context.Context, domain.EmailJob) error {
		called = true
		return nil
	})

	err := send(context.Background(), domain.EmailJob{Prospect: prospect("Acme", "a@acme.test")})

	if err == nil || called {
		t.Fatalf("err=%v called=%v, want error and no send", err, called)
	}
}
