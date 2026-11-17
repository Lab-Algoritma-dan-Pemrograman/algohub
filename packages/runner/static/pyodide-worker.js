// matplotlib di Pyodide default ke wasm_backend yang butuh `document`. Di Web
// Worker tidak ada DOM -> "ImportError: cannot import name 'document' from 'js'".
// Agg murni CPU, aman di worker; gambarnya dikirim balik sebagai PNG base64.
const MPL_SETUP = `
import os
os.environ['MPLBACKEND'] = 'Agg'
try:
    import matplotlib.pyplot as _plt
    _plt.show = lambda *a, **k: None
except Exception:
    pass
`;

const CAPTURE_PLOTS = `
import io, base64

def _capture_plots():
    try:
        import matplotlib.pyplot as plt
    except Exception:
        return []
    imgs = []
    for n in plt.get_fignums():
        b = io.BytesIO()
        plt.figure(n).savefig(b, format='png', dpi=110, bbox_inches='tight', facecolor='white')
        imgs.append('data:image/png;base64,' + base64.b64encode(b.getvalue()).decode())
        b.close()
    plt.close('all')
    return imgs

_capture_plots()
`;

const FLAG_CANCEL = 2;
const quiet = { messageCallback: () => {}, errorCallback: () => {} };

let pyodide = null;
let meta = null;
let data = null;
let runId = null;
let batchLines = null;
let waitingForInput = false;

const errorMessage = (e) => (e && e.message) || String(e);

function createWriter() {
  const decoder = new TextDecoder();
  return {
    write(bytes) {
      const text = decoder.decode(bytes, { stream: true });
      if (runId !== null && text) self.postMessage({ type: 'write', id: runId, data: text });
      return bytes.length;
    },
  };
}

async function loadFromMirrors(indexUrls) {
  let lastError = null;
  for (const indexURL of indexUrls) {
    try {
      self.importScripts(`${indexURL}pyodide.js`);
      return await self.loadPyodide({ indexURL });
    } catch (e) {
      lastError = e;
    }
  }
  throw lastError;
}

function readInteractive() {
  self.postMessage({ type: 'need-input', id: runId });
  Atomics.wait(meta, 0, 0);
  if (Atomics.load(meta, 0) === FLAG_CANCEL) throw new Error('Program dibatalkan');
  const line = new TextDecoder().decode(data.slice(0, Atomics.load(meta, 1)));
  Atomics.store(meta, 1, 0);
  Atomics.store(meta, 0, 0);
  return line;
}

function readStdin() {
  if (batchLines === null) return readInteractive();
  // Satu baris per panggilan supaya sisa input yang tidak dibaca tidak
  // tertinggal di buffer sys.stdin dan terbawa ke run berikutnya.
  const line = batchLines.shift();
  if (line === undefined) waitingForInput = true;
  return line ?? null;
}

async function init({ sab, indexUrls }) {
  try {
    pyodide = await loadFromMirrors(indexUrls);
    if (sab) {
      meta = new Int32Array(sab, 0, 2);
      data = new Uint8Array(sab, 8);
    }
    pyodide.setStdout(createWriter());
    pyodide.setStderr(createWriter());
    pyodide.setStdin({ stdin: readStdin });
    self.postMessage({ type: 'ready' });
  } catch (e) {
    self.postMessage({ type: 'init-error', error: errorMessage(e) });
  }
}

async function capturePlots() {
  try {
    const result = await pyodide.runPythonAsync(CAPTURE_PLOTS);
    return result && result.toJs ? Array.from(result.toJs()) : [];
  } catch {
    return [];
  }
}

async function run({ id, code, stdin }) {
  runId = id;
  waitingForInput = false;
  batchLines = stdin === undefined && meta ? null : (stdin ?? '').match(/[^\n]*\n|[^\n]+$/g) || [];
  let error = null;
  try {
    // Error impor sengaja diabaikan: paket yang tidak ada muncul sebagai
    // ModuleNotFoundError saat kode dijalankan. MPL_SETUP wajib setelah paket
    // dimuat karena MPLBACKEND harus diset sebelum matplotlib.pyplot diimpor.
    await pyodide.loadPackagesFromImports(code, quiet).catch(() => {});
    await pyodide.runPythonAsync(MPL_SETUP);
    await pyodide.runPythonAsync(code);
  } catch (e) {
    error = errorMessage(e);
  }
  const images = await capturePlots();
  self.postMessage({ type: 'exit', id, error, waitingForInput, images });
  runId = null;
}

self.onmessage = (event) => {
  const message = event.data;
  if (message.type === 'init') init(message);
  else if (message.type === 'run') run(message);
};
