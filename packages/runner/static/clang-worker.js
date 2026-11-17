/*
 * Copyright 2020 WebAssembly Community Group participants
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

/* global API */
self.importScripts('vendor/shared.js' + self.location.search);

let api = null;
let currentModule = null;
let lastCompiledCode = '';
let runId = null;
let compiling = false;
let compileOutput = '';

const errorMessage = (e) => (e && e.message) || String(e);

function hostWrite(data) {
  if (runId === null || !data) return;
  if (compiling) compileOutput += data;
  else self.postMessage({ type: 'write', id: runId, data });
}

function hostNeedInput() {
  self.postMessage({ type: 'need-input', id: runId });
}

function init({ sab, assets }) {
  const modules = { clang: assets.clang, lld: assets.lld, memfs: assets.memfs };
  api = new API({
    compileStreaming: (name) => Promise.resolve(modules[name]),
    readBuffer: () => Promise.resolve(assets.sysroot),
    hostWrite,
    hostNeedInput,
    sab: sab || null,
  });
  // Module sudah dikompilasi di main thread; mengisi cache mencegah log
  // "Fetching and compiling" ikut masuk ke diagnostik compile.
  api.moduleCache.clang = modules.clang;
  api.moduleCache.lld = modules.lld;
  api.ready.then(
    () => self.postMessage({ type: 'ready' }),
    (e) => self.postMessage({ type: 'init-error', error: errorMessage(e) }),
  );
}

async function compile(id, code) {
  if (code === lastCompiledCode && currentModule) return true;
  compiling = true;
  compileOutput = '';
  try {
    await api.compile({ input: 'main.c', contents: code, obj: 'main.o' });
    await api.link('main.o', 'main.wasm');
    currentModule = await WebAssembly.compile(api.memfs.getFileContents('main.wasm'));
    lastCompiledCode = code;
    return true;
  } catch (e) {
    currentModule = null;
    lastCompiledCode = '';
    self.postMessage({ type: 'compile-error', id, output: compileOutput, error: errorMessage(e) });
    return false;
  } finally {
    compiling = false;
  }
}

async function run({ id, code, stdin }) {
  runId = id;
  try {
    if (!(await compile(id, code))) return;
    const memfs = api.memfs;
    const interactive = stdin === undefined && !!memfs.sab;
    memfs.sabActive = interactive;
    memfs.sabPending = null;
    memfs.sabPendingPos = 0;
    memfs.sabWaiting = false;
    // Saat SAB aktif stdin harus kosong supaya program menunggu input asli.
    // Mode batch memakai "\n" bila kosong: tanpa input, getchar() langsung EOF
    // dan pola `while (getchar() != '\n')` yang umum di materi jadi infinite loop.
    memfs.setStdinStr(interactive ? '' : stdin || '\n');
    self.postMessage({ type: 'started', id });
    let error = null;
    try {
      await api.run(currentModule, 'main.wasm');
    } catch (e) {
      error = errorMessage(e);
    }
    memfs.sabActive = false;
    self.postMessage({ type: 'exit', id, error, waitingForInput: memfs.waitingForInput });
  } finally {
    runId = null;
  }
}

self.onmessage = (event) => {
  const message = event.data;
  if (message.type === 'init') init(message);
  else if (message.type === 'run') run(message);
};
