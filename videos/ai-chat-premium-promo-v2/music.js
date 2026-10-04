// Score for ai-chat-premium-promo-v2 (string sections after the Austerlitz score recipe, timing from timing.js).
// Based on the v1 score: warm, dark strings, a soft pad of low felt piano and a round bass.
// No sound effects of any kind and nothing in the high register: every voice runs through a low-pass,
// so nothing reads as a bleep. Chords change with the picture (one per section). Built from the
// ai-chat-teaser instruments, synthesized from scratch, no samples, no licensed audio.
// Synthesized from scratch, no samples, no licensed audio.
// Usage: node music.js <out.wav>
const fs = require('fs');
const TIMING = require('./timing.js');
const SR = 44100, DUR = TIMING.DUR, N = Math.ceil(DUR * SR);
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

/* ---------- string section (Battle-of-Austerlitz-Film tools/mix.py recipe) ----------
   Four detuned voices per note (-9/-3/+3/+9 cents), random start phase, 0.3 % vibrato at ~5.2 Hz that
   fades in, band-limited harmonics, a gentle per-voice low-pass, slow attack, long release, notes spread
   across the stereo field. gainFn shapes swells over absolute time. */
function section(t0, t1, notes, gainFn, { att: A = 1.6, rel = 2.2, lp = 2200 } = {}) {
  const i0 = Math.round(t0 * SR), i1 = Math.min(N, Math.round((t1 + rel) * SR)), a = 1 - Math.exp(-2 * Math.PI * lp / SR);
  const vs = []; notes.forEach((m, k) => [-9, -3, 3, 9].forEach((c, d) => {
    const f = mtof(m) * Math.pow(2, c / 1200), kmax = Math.max(1, Math.min(10, Math.floor(3800 / f)));
    vs.push({ f, kmax, ph: (rnd() + 1) / 2, vr: 5.0 + .4 * ((rnd() + 1) / 2), vp: (rnd() + 1) * 3, pan: clamp(-.7 + 1.4 * (k / Math.max(1, notes.length - 1)) + (d - 1.5) * .05, -.8, .8), y: 0 });
  }));
  for (const v of vs) { v.gl = Math.cos((v.pan + 1) * Math.PI / 4); v.gr = Math.sin((v.pan + 1) * Math.PI / 4); }
  const norm = 1 / Math.sqrt(vs.length);
  for (let i = i0; i < i1; i++) {
    const ta = i / SR, tt = ta - t0;
    const env = Math.min(1, tt / A) ** 2 * (ta > t1 ? Math.max(0, 1 - (ta - t1) / rel) : 1) * gainFn(ta);
    if (env <= 1e-5) continue;
    const vib = Math.min(1, tt / .4);
    let l = 0, r = 0;
    for (const v of vs) {
      v.ph += v.f * (1 + .003 * vib * ST[(((ta * v.vr + v.vp) % 1) * TS) | 0]) / SR;
      let x = 0; for (let h = 1; h <= v.kmax; h++) { const q = v.ph * h; x += ST[((q - Math.floor(q)) * TS) | 0] / h; }
      v.y += a * (x - v.y); l += v.y * v.gl; r += v.y * v.gr;
    }
    const g = env * norm;
    B.pad[0][i] += l * g; B.pad[1][i] += r * g; B.rev[0][i] += l * g * .45; B.rev[1][i] += r * g * .45;
  }
}

/* ---------- form, all from timing.js: intro 6 beats, eight 2-bar sections, 2-bar end card ---------- */
const { BEAT, INTRO, SL, OUT, HEADLINE, S } = TIMING;
// arc (Austerlitz: minor until the light breaks, then major): darker colours first, home to F major at the end
const CH = [
  { b: 26, v: [50, 53, 57, 60] }, { b: 34, v: [50, 53, 57, 62] }, { b: 31, v: [50, 53, 58, 62] }, { b: 33, v: [52, 55, 57, 60] },
  { b: 29, v: [53, 57, 60, 64] }, { b: 26, v: [50, 53, 57, 64] }, { b: 34, v: [50, 55, 57, 62] }, { b: 36, v: [52, 55, 60, 62] },
];
const swell = t0 => t => { const h = t0 + HEADLINE; return 1 + .7 * (t < h ? clamp((t - (h - 1.8)) / 1.8) ** 2 : Math.exp(-(t - h) / 1.4)); };
// intro: a low note on each word (on the beat), a warm chord on "Just ask."
[0, 1, 2, 3].forEach(k => { const t = .3 + k * BEAT; [38, 50, 53].forEach((m, j) => piano(t + j * .02, m + [0, -2, -4, -2][k], .2, .8, null, .5)); });
[26, 38, 45, 50, 53, 57].forEach((m, j) => piano(4 * BEAT + j * .05, m, .25, 1.8, null, .55));
section(.2, INTRO + .4, [38, 50, 53, 57], () => .018, { att: 2.5 });
// sections
for (let i = 0; i < 8; i++) {
  const c = CH[i], t0 = S(i), t1 = t0 + SL, sw = swell(t0);
  section(t0, t1, [c.b + 12, ...c.v], t => .03 * sw(t), { att: 1.2, rel: 1.6, lp: 2000 + 400 * (i / 7) });
  voice(B.bass, t0, SL + .9, tt => (sn(mtof(c.b) * tt) + .12 * sn(mtof(c.b + 12) * tt)) * att(tt, .4) * Math.exp(-tt * .2), .2);
  for (let q = 0; q < 8; q++) {                                                  // one low felt-piano note per beat
    const t = t0 + q * BEAT, m = c.v[[0, 2, 1, 3, 0, 2, 1, 2][q]];
    piano(t, m, (q % 4 === 0 ? .17 : .12) * (1 + .05 * rnd()), BEAT * 1.1, q % 2 ? .2 : -.2, .45);
  }
  [c.b + 12, c.v[0], c.v[1], c.v[2]].forEach((m, j) => piano(t0 + HEADLINE + j * .04, m, .19, 1.7, null, .55));   // headline lands
}
// end card: home to F major on "Just ask." and let it ring
const END = OUT + 2 * BEAT;
[29, 41, 48, 53, 57, 60, 64].forEach((m, j) => piano(END + j * .06, m, j < 2 ? .3 : .24, 4, null, .6));
section(OUT, DUR - 1.2, [41, 53, 57, 60, 64, 67], t => .022 + .02 * clamp((t - (END - 1.2)) / 1.2), { att: 1.0, rel: 1.2, lp: 2600 });
voice(B.bass, OUT, DUR - OUT, tt => sn(mtof(29) * tt) * att(tt, .5) * Math.exp(-tt * .22), .2);

/* ---------- low-pass every bus (no brightness, nothing chippy), reverb, master ---------- */
function lowpass(buf, fc) { const a = 1 - Math.exp(-2 * Math.PI * fc / SR); for (const ch of buf) { let y1 = 0, y2 = 0; for (let n = 0; n < N; n++) { y1 += a * (ch[n] - y1); y2 += a * (y1 - y2); ch[n] = y2; } } }
lowpass(B.keys, 2600); lowpass(B.pad, 3400); lowpass(B.rev, 2000);   // the reverb tail darkens (Austerlitz IR)
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
