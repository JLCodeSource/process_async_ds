package main

type mockAsyncProcessor struct {
	Env   *env
	Files *[]file
}

func (m mockAsyncProcessor) getFiles() *[]file {
	return m.Files
}

func (m mockAsyncProcessor) getEnv() *env {
	return m.Env
}

func (m mockAsyncProcessor) setEnv(_ *env) {
	//m.Env = env
}

func (m mockAsyncProcessor) setFiles() {
}

func (m mockAsyncProcessor) processFiles() {
}

func (m mockAsyncProcessor) parseSourceFile() []string {
	return []string{}
}

func (m mockAsyncProcessor) parseLine(_ string) file {
	var f file
	return f
}
