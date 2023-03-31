package main

// mockAsyncProcessor

type mockAsyncProcessor struct {
	env   *env
	files []file
}

func (m mockAsyncProcessor) getFiles() []file {
	return m.files
}

func (m mockAsyncProcessor) getEnv() *env {
	return m.env
}

func (m mockAsyncProcessor) setEnv(_ *env) {
	//m.Env = env
}

func (m mockAsyncProcessor) setFiles() {
}

func (m mockAsyncProcessor) processFiles() {
}
