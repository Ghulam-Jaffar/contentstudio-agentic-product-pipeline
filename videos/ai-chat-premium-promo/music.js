// Score for ai-chat-premium-promo: warm, dark strings, a soft pad of low felt piano and a round bass.
// No sound effects of any kind and nothing in the high register: every voice runs through a low-pass,
// so nothing reads as a bleep. Chords change with the picture (one per section). Built from the
// ai-chat-teaser instruments, synthesized from scratch, no samples, no licensed audio.
// Synthesized from scratch, no samples, no licensed audio.
// Usage: node music.js <out.wav>
const fs = require('fs');
const SR = 44100, DUR = 66, N = Math.ceil(DUR * SR);
const mk = () => [new Float32Array(N), new Float32Array(N)];
const B = { keys: mk(), pad: mk(), bass: mk(), rev: mk(), dly: mk() };
let seed = 11; const rnd = () => { seed = (seed * 16807) % 2147483647; return seed / 2147483647 * 2 - 1; };
const clamp = (x, a = 0, b = 1) => Math.max(a, Math.min(b, x));
const mtof = m => 440 * Math.pow(2, (m - 69) / 12);
const TS = 8192, ST = new Float32Array(TS); for (let i = 0; i < TS; i++) ST[i] = Math.sin(2 * Math.PI * i / TS);
const sn = ph => ST[((ph - Math.floor(ph)) * TS) | 0];
const att = (tt, a) => (tt < a ? tt / a : 1);

/* ---------- felt piano: inharmonic partials, two strings per note, two-stage decay ---------- */
function piano(t, m, vel, hold = 1.2, pan = null, rs = .4, ds = 0) {
  t += rnd() * .007; vel = clamp(vel * (1 + rnd() * .06), .05, 1);
  const f = mtof(m), Binh = .00032, base = 3.2 * Math.pow(2, -(m - 50) / 22);
  const P = []; for (let n = 1; n <= 14; n++) { const fn = n * f * Math.sqrt(1 + Binh * n * n); if (fn > 7000) break; P.push({ fn, a: vel * Math.pow(n, -1.35) * Math.exp(-(n - 1) * (1.05 - vel * .55)), tau: base / (1 + .45 * (n - 1)) }); }
  pan = pan ?? clamp((m - 60) / 30, -.6, .6);
  const gl = Math.cos((pan + 1) * Math.PI / 4), gr = Math.sin((pan + 1) * Math.PI / 4);
  const i0 = Math.round(t * SR), len = Math.round((hold + 1.2) * SR);
  const st = P.map(p => ({ ph1: 0, ph2: .25, d1: p.fn / SR, d2: p.fn * Math.pow(2, .9 / 1200) / SR, k1: Math.exp(-1 / (p.tau * .22 * SR)), k2: Math.exp(-1 / (p.tau * SR)), e1: p.a * .55, e2: p.a * .45 }));
  let hlp = 0;
  for (let j = 0; j < len; j++) {
    const i = i0 + j; if (i < 0 || i >= N) continue;
    const tt = j / SR; let s = 0;
    for (const q of st) {
      s += (q.e1 + q.e2) * (ST[((q.ph1 - Math.floor(q.ph1)) * TS) | 0] + ST[((q.ph2 - Math.floor(q.ph2)) * TS) | 0]) * .5;
      q.ph1 += q.d1; q.ph2 += q.d2; q.e1 *= q.k1; q.e2 *= q.k2;
    }
    const damp = tt > hold ? Math.exp(-(tt - hold) / .18) : 1;
    if (tt < .03) { hlp += .08 * (rnd() - hlp); s += hlp * vel * .5 * (1 - tt / .03); }
    const v = s * att(tt, .006) * damp * .32;
    B.keys[0][i] += v * gl; B.keys[1][i] += v * gr;
    B.rev[0][i] += v * gl * rs; B.rev[1][i] += v * gr * rs;
    if (ds) { B.dly[0][i] += v * gl * ds; B.dly[1][i] += v * gr * ds; }
  }
}

/* ---------- warm strings: slow bow, vibrato, dark tone ---------- */
function strings(t0, t1, notes, gain, bright = 1.6, rel = 1.4) {
  const H = 7, i0 = Math.round(t0 * SR), i1 = Math.min(N, Math.round((t1 + rel) * SR));
  const W = []; for (let h = 1; h <= H; h++) W.push(Math.exp(-h / bright) / h);
  const vs = []; notes.forEach((m, k) => [-6, 5].forEach((dc, d) => vs.push({ f: mtof(m) * Math.pow(2, dc / 1200), ph: (k * .13 + d * .37) % 1, vr: 4.6 + k * .31 + d * .2, left: !d })));
  for (let i = i0; i < i1; i++) {
    const tt = (i - i0) / SR, ta = i / SR;
    const env = Math.min(1, tt / .9) ** 2 * (ta > t1 ? Math.max(0, 1 - (ta - t1) / rel) : 1);
    if (env <= 0) continue;
    let l = 0, r = 0;
    for (const v of vs) {
      v.ph += v.f * (1 + .0018 * ST[(((ta * v.vr) % 1) * TS) | 0]) / SR; let s = 0;
      for (let h = 1; h <= H; h++) { const x = v.ph * h; s += ST[((x - Math.floor(x)) * TS) | 0] * W[h - 1]; }
      if (v.left) l += s; else r += s;
    }
    const g = env * gain;
    B.pad[0][i] += (l * .8 + r * .2) * g; B.pad[1][i] += (r * .8 + l * .2) * g;
    B.rev[0][i] += l * g * .35; B.rev[1][i] += r * g * .35;
  }
}

/* ---------- voice writer ---------- */
function voice(bus, t0, dur, fn, gain, pan = 0, rs = 0, ds = 0) {
  const l = Math.cos((pan + 1) * Math.PI / 4) * gain, r = Math.sin((pan + 1) * Math.PI / 4) * gain;
  const i0 = Math.max(0, Math.round(t0 * SR)), i1 = Math.min(N, Math.round((t0 + dur) * SR));
  for (let i = i0; i < i1; i++) {
    const v = fn((i - t0 * SR) / SR, i / SR);
    bus[0][i] += v * l; bus[1][i] += v * r;
    if (rs) { B.rev[0][i] += v * l * rs; B.rev[1][i] += v * r * rs; }
    if (ds) { B.dly[0][i] += v * l * ds; B.dly[1][i] += v * r * ds; }
  }
}
function bassNote(t, m) {       // round, upright-ish: sine body, a little 2nd, soft attack
  const f = mtof(m);
  voice(B.bass, t, .3, tt => (sn(f * tt) + .22 * sn(2 * f * tt)) * att(tt, .015) * Math.exp(-tt * 4.5) * (tt > .26 ? (.3 - tt) / .04 : 1), .27);
}

/* ---------- form: intro 0-4.6, eight 7 s sections, outro 60.6-66 (matches source.html) ---------- */
const INTRO = 4.6, SL = 7, OUT = INTRO + 8 * SL, S = i => INTRO + i * SL;
// one chord per section, all in F: Fmaj9 Dm9 Bbmaj9 Gm9 Am7 Dm9 Bbmaj9 C(add9) -> F
const CH = [
  { b: 29, v: [53, 57, 60, 64] }, { b: 26, v: [50, 53, 57, 60] }, { b: 34, v: [50, 53, 57, 62] }, { b: 31, v: [50, 53, 58, 62] },
  { b: 33, v: [52, 55, 57, 60] }, { b: 26, v: [50, 53, 57, 64] }, { b: 34, v: [50, 55, 57, 62] }, { b: 36, v: [52, 55, 60, 62] },
];
const FIN = { b: 29, v: [53, 57, 60, 64, 67] };
// intro: a low chord on each word, a fuller one on "Just ask."
[0, 1, 2, 3].forEach(k => { const t = .25 + k * .62; [41, 53, 57].forEach((m, j) => piano(t + j * .02, m + (k % 2 ? -2 : 0), .2, .7, null, .5)); });
[29, 41, 48, 53, 57, 60].forEach((m, j) => piano(2.75 + j * .05, m, .26, 1.8, null, .55));
strings(2.6, INTRO + .4, [41, 53, 57, 60], .02, 1.1, 1.2);
// sections
for (let i = 0; i < 8; i++) {
  const c = CH[i], t0 = S(i), t1 = t0 + SL;
  strings(t0, t1, [c.b + 12, ...c.v], .026, 1.15, 1.4);                         // dark, warm bed
  voice(B.bass, t0, SL + .8, tt => (sn(mtof(c.b) * tt) + .12 * sn(mtof(c.b + 12) * tt)) * att(tt, .4) * Math.exp(-tt * .22), .2);
  for (let q = 0; q < 9; q++) {                                                // a soft low pulse, quarter notes at 80 BPM
    const t = t0 + q * .75, m = c.v[[0, 2, 1, 3, 0, 2, 1, 2, 0][q]];
    piano(t, m, q % 4 === 0 ? .17 : .12, .9, q % 2 ? .2 : -.2, .45);
  }
  [c.b + 12, c.v[0], c.v[1], c.v[2]].forEach((m, j) => piano(t0 + 5.3 + j * .04, m, .18, 1.6, null, .55));   // headline lands
}
// outro: "Just ask." resolves home and rings out
[29, 41, 48, 53, 57, 60, 64].forEach((m, j) => piano(OUT + 1.5 + j * .06, m, j < 2 ? .3 : .24, 3.6, null, .6));
strings(OUT, DUR - 1.4, [FIN.b + 12, ...FIN.v], .032, 1.2, 1.4);
voice(B.bass, OUT, DUR - OUT, tt => sn(mtof(29) * tt) * att(tt, .5) * Math.exp(-tt * .25), .2);

/* ---------- low-pass every bus (no brightness, nothing chippy), reverb, master ---------- */
function lowpass(buf, fc) { const a = 1 - Math.exp(-2 * Math.PI * fc / SR); for (const ch of buf) { let y1 = 0, y2 = 0; for (let n = 0; n < N; n++) { y1 += a * (ch[n] - y1); y2 += a * (y1 - y2); ch[n] = y2; } } }
lowpass(B.keys, 2600); lowpass(B.pad, 3200); lowpass(B.rev, 2400);
function freeverb(inL, inR) {
  const C = [1116, 1188, 1277, 1356, 1422, 1491, 1557, 1617], A = [556, 441, 341, 225], fb = .87, damp = .42, pre = Math.round(.02 * SR);
  const out = [new Float32Array(N), new Float32Array(N)];
  [inL, inR].forEach((inp, ch) => {
    const sp = ch * 23, combs = C.map(c => ({ b: new Float32Array(c + sp), i: 0, fs: 0 })), aps = A.map(a => ({ b: new Float32Array(a + sp), i: 0 }));
    for (let n = 0; n < N; n++) {
      const x = (n >= pre ? inp[n - pre] : 0) * .015; let o = 0;
      for (const c of combs) { const y = c.b[c.i]; c.fs = y * (1 - damp) + c.fs * damp; c.b[c.i] = x + c.fs * fb; if (++c.i >= c.b.length) c.i = 0; o += y; }
      for (const a of aps) { const bo = a.b[a.i]; const y = -o + bo; a.b[a.i] = o + bo * .5; if (++a.i >= a.b.length) a.i = 0; o = y; }
      out[ch][n] = o;
    }
  });
  return out;
}
const rev = freeverb(B.rev[0], B.rev[1]);
const L = new Float32Array(N), R = new Float32Array(N); let peak = 0;
for (let n = 0; n < N; n++) {
  const t = n / SR, fade = clamp(t / .4) * (t > DUR - 1.4 ? Math.max(0, (DUR - t) / 1.4) : 1);
  for (let ch = 0; ch < 2; ch++) {
    const v = B.keys[ch][n] + B.pad[ch][n] * 1.1 + B.bass[ch][n] + rev[ch][n] * 3.0;
    const o = Math.tanh(v * 1.02) * fade; (ch ? R : L)[n] = o; peak = Math.max(peak, Math.abs(o));
  }
}
const g = .89 / peak, buf = Buffer.alloc(44 + N * 4);
buf.write('RIFF', 0); buf.writeUInt32LE(36 + N * 4, 4); buf.write('WAVE', 8); buf.write('fmt ', 12);
buf.writeUInt32LE(16, 16); buf.writeUInt16LE(1, 20); buf.writeUInt16LE(2, 22); buf.writeUInt32LE(SR, 24);
buf.writeUInt32LE(SR * 4, 28); buf.writeUInt16LE(4, 32); buf.writeUInt16LE(16, 34); buf.write('data', 36); buf.writeUInt32LE(N * 4, 40);
for (let n = 0; n < N; n++) { buf.writeInt16LE(Math.round(clamp(L[n] * g, -1, 1) * 32767), 44 + n * 4); buf.writeInt16LE(Math.round(clamp(R[n] * g, -1, 1) * 32767), 46 + n * 4); }
fs.writeFileSync(process.argv[2] || 'music.wav', buf);
console.log('wrote', process.argv[2] || 'music.wav', DUR + 's');
