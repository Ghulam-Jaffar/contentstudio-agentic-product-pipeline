// Score for ai-chat-style-test: felt piano, warm strings and a soft round bass, built from the
// ai-chat-teaser instruments. No whooshes, chimes, sparkles, risers, impacts or UI clicks:
// the only accents are piano notes on the picture's moments (images develop, headline lands).
// Synthesized from scratch, no samples, no licensed audio.
// Usage: node music.js <out.wav>
const fs = require('fs');
const SR = 44100, DUR = 12, N = Math.ceil(DUR * SR);
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

/* ---------- score: 72 BPM, Fmaj9 -> Am7 -> Dm9 -> Bbmaj9, one chord per bar ---------- */
// picture moments (12 s cut): send 2.64, images develop 4.7 / 5.15 / 5.6 / 6.05, push-in 9.0, headline 9.6
const BEAT = 60 / 72, BAR = BEAT * 4;
const CH = [{ lh: [29, 41], rh: [53, 57, 60, 64, 67] }, { lh: [33, 45], rh: [52, 57, 60, 64, 67] }, { lh: [26, 38], rh: [50, 53, 57, 60, 64] }, { lh: [22, 34], rh: [50, 53, 57, 62, 65] }];
const chordAt = t => Math.min(3, Math.floor(t / BAR));
// strings bed, one chord per bar, the last one held into the end
for (let b = 0; b < 4; b++) { const c = CH[b], t0 = b * BAR, t1 = b === 3 ? DUR - 1.6 : t0 + BAR;
  strings(t0, t1, [c.lh[1] + 12, ...c.rh.slice(0, 4)], b === 0 ? .022 : .028, 1.5, 1.2); }
// left hand + soft bass on each bar
for (let b = 0; b < 4; b++) { const c = CH[b], t = b * BAR;
  c.lh.forEach((m, k) => piano(t + k * .02, m, .3, BAR * .9, -.15, .45));
  voice(B.bass, t, BAR, tt => (sn(mtof(c.lh[0] + 12) * tt) + .15 * sn(mtof(c.lh[0] + 24) * tt)) * att(tt, .25) * Math.exp(-tt * .55), .16); }
// right hand: a slow, sparse broken chord in quarter notes, quiet under the typing
const PAT = [0, 2, 4, 3, 1, 3, 2, 4];
for (let s = 0; s * BEAT < 9.4; s++) { const t = s * BEAT, c = CH[chordAt(t)];
  piano(t, c.rh[PAT[s % 8]] + 12, s % 4 === 0 ? .2 : .15, BEAT * 1.6, s % 2 ? .25 : -.25, .5, .12); }
// accents on the picture: a soft high note as each image develops (a melody, not a chime)
[[4.7, 81], [5.15, 79], [5.6, 76], [6.05, 72]].forEach(([t, m]) => piano(t, m, .26, 1.6, .2, .6, .2));
piano(9.0, 77, .24, 1.4, -.1, .6, .2);                       // push-in on the first image
// headline lands: a gently rolled Fmaj9 that rings out to the end
[29, 41, 53, 57, 60, 64, 67, 72].forEach((m, k) => piano(9.6 + k * .045, m, k < 2 ? .34 : .28, 2.4, null, .6));
strings(9.6, DUR - 1.0, [53, 57, 60, 64, 69], .03, 1.7, 1.0);

/* ---------- reverb, delay, master ---------- */
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
const D = Math.round(BEAT * .75 * SR), dl = [new Float32Array(N), new Float32Array(N)];
let dlL = 0, dlR = 0;
for (let n = 0; n < N; n++) {
  const bl = n >= D ? dl[1][n - D] : 0, br = n >= D ? dl[0][n - D] : 0;
  dlL += .3 * (bl - dlL); dlR += .3 * (br - dlR);
  dl[0][n] = B.dly[0][n] + dlL * .3; dl[1][n] = B.dly[1][n] + dlR * .3;
}
const L = new Float32Array(N), R = new Float32Array(N); let peak = 0;
for (let n = 0; n < N; n++) {
  const t = n / SR, fade = clamp(t / .6) * (t > DUR - 1.2 ? Math.max(0, (DUR - t) / 1.2) : 1);
  for (let ch = 0; ch < 2; ch++) {
    const v = B.keys[ch][n] + B.pad[ch][n] + B.bass[ch][n] + rev[ch][n] * 2.8 + (dl[ch][n] - B.dly[ch][n]) * .7;
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
