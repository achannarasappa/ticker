package snaptrade

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
)

var _ = Describe("Store", func() {

	var fs afero.Fs

	BeforeEach(func() {
		fs = afero.NewMemMapFs()
	})

	It("should return ok=false when no secret is stored", func() {
		secret, ok, err := NewStore(fs, "/data").GetUserSecret("ben")

		Expect(err).NotTo(HaveOccurred())
		Expect(ok).To(BeFalse())
		Expect(secret).To(BeEmpty())
	})

	It("should persist and read back a user secret", func() {
		store := NewStore(fs, "/data")

		Expect(store.SaveUserSecret("ben", "secret-123")).To(Succeed())

		secret, ok, err := store.GetUserSecret("ben")

		Expect(err).NotTo(HaveOccurred())
		Expect(ok).To(BeTrue())
		Expect(secret).To(Equal("secret-123"))

		exists, _ := afero.Exists(fs, "/data/ticker/snaptrade.json")
		Expect(exists).To(BeTrue())
	})

	It("should round-trip account preferences", func() {
		store := NewStore(fs, "/data")

		prefs := Preferences{HiddenAccounts: []string{"acct-2"}, DefaultAccountID: "acct-1"}
		Expect(store.SavePreferences(prefs)).To(Succeed())

		out, err := store.GetPreferences()
		Expect(err).NotTo(HaveOccurred())
		Expect(out.DefaultAccountID).To(Equal("acct-1"))
		Expect(out.IsHidden("acct-2")).To(BeTrue())
		Expect(out.IsHidden("acct-1")).To(BeFalse())
	})

	It("should preserve user secrets and preferences independently", func() {
		store := NewStore(fs, "/data")

		Expect(store.SaveUserSecret("ben", "secret-123")).To(Succeed())
		Expect(store.SavePreferences(Preferences{DefaultAccountID: "acct-1"})).To(Succeed())

		secret, ok, _ := store.GetUserSecret("ben")
		Expect(ok).To(BeTrue())
		Expect(secret).To(Equal("secret-123"))

		prefs, _ := store.GetPreferences()
		Expect(prefs.DefaultAccountID).To(Equal("acct-1"))
	})

	It("should return empty preferences when none are stored", func() {
		prefs, err := NewStore(fs, "/data").GetPreferences()

		Expect(err).NotTo(HaveOccurred())
		Expect(prefs.HiddenAccounts).To(BeEmpty())
		Expect(prefs.DefaultAccountID).To(BeEmpty())
	})

	It("should preserve other users when saving", func() {
		store := NewStore(fs, "/data")

		Expect(store.SaveUserSecret("ben", "secret-ben")).To(Succeed())
		Expect(store.SaveUserSecret("alice", "secret-alice")).To(Succeed())

		benSecret, _, _ := store.GetUserSecret("ben")
		aliceSecret, _, _ := store.GetUserSecret("alice")

		Expect(benSecret).To(Equal("secret-ben"))
		Expect(aliceSecret).To(Equal("secret-alice"))
	})
})
