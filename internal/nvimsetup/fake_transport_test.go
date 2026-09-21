package nvimsetup

// fakeTransport is an in-memory transport.Transport for testing Ensure's
// strategy logic without touching docker/ssh. Each field is a canned
// response; a nil func means "not expected to be called in this test".
type fakeTransport struct {
	homeOut string
	homeErr error

	unameOS, unameArch string
	unameErr           error

	runFunc  func(script string) (string, error)
	copyFunc func(localPath, remotePath string) error
	mkdirErr error
}

func (f *fakeTransport) RunInteractive(env map[string]string, args ...string) error {
	panic("fakeTransport: RunInteractive not expected in these tests")
}

func (f *fakeTransport) Run(script string) (string, error) {
	if f.runFunc == nil {
		return "", nil
	}
	return f.runFunc(script)
}

func (f *fakeTransport) CopyTo(localPath, remotePath string) error {
	if f.copyFunc == nil {
		return nil
	}
	return f.copyFunc(localPath, remotePath)
}

func (f *fakeTransport) MkdirAll(remotePath string) error { return f.mkdirErr }

func (f *fakeTransport) Home() (string, error) { return f.homeOut, f.homeErr }

func (f *fakeTransport) Uname() (string, string, error) {
	return f.unameOS, f.unameArch, f.unameErr
}
