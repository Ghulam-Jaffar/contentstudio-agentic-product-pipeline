// Original score for ai-chat-teaser. Same arrangement and sound design as the promo's score
// (120 BPM groove, a whoosh on every cut, clicks, success chimes, sparkles, risers,
// the carousel drop, montage stabs, the final hit), but the tune is carried by acoustic-style
// instruments: a felt piano instead of the synth arp, warm strings instead of the synth pad,
// and a rounder bass. Synthesized from scratch, no samples, no licensed audio.
// Cues come from the teaser's CUT table in source.html, so the music follows the edit.
//
// Usage: node music.js <out.wav>

const fs = require('fs');
const SR = 44100;

/* ---------- the teaser's edit (keep in sync with CUT in source.html) ---------- */
const CUT = [['s0', 0, 6.6, 1.25], ['s2', 13.3, 26.2, 1.65], ['s3', 25.8, 35.2, 1.65], ['s4', 34.8, 43, 1.6], ['s5', 42.6, 54.4, 1.6], ['s6', 51.6, 64.6, 1.65], ['s7', 64.2, 75.4, 1.7], ['s8', 75, 85.4, 1.7], ['s9', 85, 93.2, 1.6], ['s11', 101.2, 106.4, 1], ['s12', 106, 113, 1]];
let acc = 0; const SCH = {};
for (const [id, a, b, sp] of CUT) { SCH[id] = { a, b, sp, t0: acc, t1: acc + (b - a) / sp }; acc = SCH[id].t1 - .3; }
// promo time -> teaser time, or null when that moment was cut
const M = T => { for (const [id] of CUT) { const g = SCH[id]; if (T >= g.a && T <= g.b) return g.t0 + (T - g.a) / g.sp; } return null; };
const DUR = SCH.s12.t1, N = Math.ceil(DUR * SR);

const mk = () => [new Float32Array(N), new Float32Array(N)];
const B = { keys: mk(), pad: mk(), bass: mk(), drum: mk(), fx: mk(), rev: mk(), dly: mk() };
let seed = 7; const rnd = () => { seed = (seed * 16807) % 2147483647; return seed / 2147483647 * 2 - 1; };
const clamp = (x, a = 0, b = 1) => Math.max(a, Math.min(b, x));
const mtof = m => 440 * Math.pow(2, (m - 69) / 12);
const TS = 8192, ST = new Float32Array(TS); for (let i = 0; i < TS; i++) ST[i] = Math.sin(2 * Math.PI * i / TS);
const sn = ph => ST[((ph - Math.floor(ph)) * TS) | 0];
const on = (t, ranges) => ranges.some(([a, b]) => t >= a && t < b);
const att = (tt, a) => (tt < a ? tt / a : 1);

/* ---------- form (the promo's arrangement, re-timed) ---------- */
const BEAT = .5, BAR = 2, G0 = SCH.s2.t0;
const snap = x => G0 + Math.round((x - G0) / BEAT) * BEAT;
const DROP = snap(M(58.2)), BREAK = SCH.s6.t0, MONT = snap(SCH.s11.t0), END = SCH.s12.t0 + .2;
const KICK = [[G0 + BAR, BREAK], [DROP, MONT]];
const HAT = [[G0, BREAK], [DROP, MONT]];
const HAT16 = [[snap(SCH.s3.t0), BREAK], [snap(SCH.s7.t0), MONT]];
const CLAP = [[snap(SCH.s3.t0), BREAK], [DROP, MONT]];
const LEAD = [[DROP, snap(SCH.s7.t0)], [snap(SCH.s9.t0), MONT]];

// Fmaj7, Am7, Dm7, Bbmaj7 (two bars each), as in the promo
const CH = [{ lh: [41, 48], rh: [53, 57, 60, 64] }, { lh: [45, 52], rh: [57, 60, 64, 67] }, { lh: [38, 45], rh: [50, 53, 57, 60] }, { lh: [34, 41], rh: [50, 53, 57, 62] }];
const chordAt = t => (t < G0 ? 0 : Math.floor((t - G0 + 1e-6) / 4) % 4);
function level(t) {                      // how hard the piano is played through the piece
  if (t < G0) return .8;
  if (t < G0 + BAR * 2) return .7;
  if (t >= BREAK && t < DROP) return .6 + .4 * clamp((t - BREAK) / (DROP - BREAK));
  return 1;
}

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

/* ---------- voice writer for drums and sound design ---------- */
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
function bell(t, m, gain, pan = 0, decay = .9, rs = .45, ds = .2) {
  const f = mtof(m);
  voice(B.fx, t, decay * 5, tt => (sn(f * tt) + .4 * sn(2 * f * tt) * Math.exp(-tt * 3) + .18 * sn(3.01 * f * tt) * Math.exp(-tt * 6)) * Math.exp(-tt / decay) * att(tt, .003), gain, pan, rs, ds);
}
function kick(t) { voice(B.drum, t, .5, tt => { const ph = 45 * tt + 100 * (1 - Math.exp(-30 * tt)) / 30; return sn(ph) * Math.exp(-tt * 7) * att(tt, .002) + rnd() * Math.exp(-tt * 400) * .2; }, .5); }
function hat(t, g, dec, pan) { let p = 0; voice(B.drum, t, .12, tt => { const x = rnd(), y = x - p; p = x; return y * Math.exp(-tt * dec); }, g, pan, .08); }
function clap(t) {
  let lp = 0, lp2 = 0;
  voice(B.drum, t, .3, tt => { const env = (tt < .01 ? 1 : tt < .02 ? .7 : 1) * Math.exp(-tt * 18) + (tt > .02 ? Math.exp(-(tt - .02) * 14) * .6 : 0); const x = rnd(); lp += .35 * (x - lp); lp2 += .08 * (lp - lp2); return (lp - lp2) * env * 2.2; }, .13, -.1, .35);
}
function bassNote(t, m) {       // round, upright-ish: sine body, a little 2nd, soft attack
  const f = mtof(m);
  voice(B.bass, t, .3, tt => (sn(f * tt) + .22 * sn(2 * f * tt)) * att(tt, .015) * Math.exp(-tt * 4.5) * (tt > .26 ? (.3 - tt) / .04 : 1), .27);
}
function svfNoise(fcFn, ampFn) {
  let low = 0, band = 0;
  return tt => { const fc = Math.min(7000, fcFn(tt)), f = 2 * Math.sin(Math.PI * fc / SR); const x = rnd(); low += f * band; const high = x - low - .5 * band; band += f * high; return band * ampFn(tt); };
}
function whoosh(tc, gain = .12, dur = .9) {
  voice(B.fx, tc - dur * .7, dur, svfNoise(tt => 300 * Math.pow(5000 / 300, tt / dur), tt => { const x = tt / dur; return Math.pow(Math.sin(Math.PI * Math.min(1, x / .7) / 2), 2) * (x > .7 ? Math.exp(-(x - .7) * 12) : 1); }), gain, 0, .3);
}
function riser(t0, t1, gain = .16) {
  const d = t1 - t0;
  voice(B.fx, t0, d, svfNoise(tt => 200 * Math.pow(35, tt / d), tt => Math.pow(tt / d, 2) * (tt > d - .01 ? (d - tt) / .01 : 1)), gain, 0, .25);
}
function impact(t, gain = .5) {
  let lp = 0;
  voice(B.fx, t, 3.2, tt => { const ph = 28 * tt + 42 * (1 - Math.exp(-4 * tt)) / 4; const x = rnd(); lp += .05 * (x - lp); return sn(ph) * Math.exp(-tt * 1.6) * att(tt, .003) + lp * 3 * Math.exp(-tt * 7); }, gain, 0, .4);
}
function click(t, g = .07) { voice(B.fx, t, .03, tt => sn(1900 * tt) * Math.exp(-tt * 250) + rnd() * Math.exp(-tt * 700) * .3, g, .15, .05); }
function chime(t) { [84, 88, 91].forEach((m, i) => bell(t + i * .09, m, .09, -.2 + i * .2, .7)); }
function sparkle(t, g = .045) { [84, 86, 89, 91, 93, 96].forEach((m, i) => bell(t + i * .055, m, g, -.4 + i * .16, .5)); }

/* ---------- score ---------- */
// strings bed under everything, one chord per two bars
strings(0, G0, [53, 57, 60, 64, 67], .03, 1.3);
for (let t = G0; t < END; t += 4) { const c = CH[chordAt(t + .01)]; const br = t >= BREAK && t < DROP;
  strings(t, Math.min(t + 4, END), [c.lh[0] + 12, ...c.rh], br ? .028 : .034, br ? 1.4 : 1.9, .9); }
strings(END, DUR - 1.2, [41, 53, 57, 60, 64, 67], .045, 1.8, 1.4);

// piano: broken chords in eighths (the promo's arp line, played on piano), left hand on each bar
const PAT = [0, 2, 4, 6, 7, 5, 3, 1];
for (let s = 0; ; s++) {
  const t = G0 + s * .25; if (t >= MONT) break;
  const c = CH[chordAt(t)], tones = [...c.rh, ...c.rh.map(x => x + 12)], lv = level(t);
  piano(t, tones[PAT[s % 8]] + 12, (s % 4 === 0 ? .34 : s % 2 ? .22 : .27) * lv, .4, s % 2 ? .3 : -.3, .3, .18);
  if (s % 8 === 0) c.lh.forEach((m, k) => piano(t + k * .015, m, .36 * lv, 1.8));
}
// lead: a high piano melody over the drop and the skills section
for (let t = G0; t < MONT; t += 1) if (on(t, LEAD)) { const c = CH[chordAt(t)]; piano(t, c.rh[((t - G0) % 2) < .5 ? 3 : 2] + 24, .4, .9, ((t | 0) % 2) ? .35 : -.35, .5, .25); }

// drums and bass on the grid
for (let n = 0; ; n++) {
  const t = G0 + n * BEAT; if (t >= MONT) break;
  if (on(t, KICK)) { kick(t); if ((n % 4) === 1 || (n % 4) === 3) if (on(t, CLAP)) clap(t); }
  const off = t + .25;
  if (on(off, HAT)) hat(off, .06, 55, .25);
  if (on(t, HAT16)) { hat(t + .125, .025, 90, -.2); hat(t + .375, .025, 90, -.2); }
  if (on(off, KICK)) bassNote(off, CH[chordAt(off)].lh[0]);
}

// intro: a note on each word, a boom and rolled chord on "Now just ask.", riser into the chat
[77, 81, 84, 88].forEach((m, k) => piano(M(.35 + .68 * k), m, .42, 1.2, -.3 + k * .2, .5));
const ASK = M(3.2); impact(ASK, .42); [41, 53, 57, 60, 64, 67, 72].forEach((m, k) => piano(ASK + k * .03, m, .42, 2.2));
riser(M(4.1), M(5.25), .13); whoosh(G0 + .05, .16, 1.1);

// a whoosh on every scene change
['s3', 's4', 's5', 's6', 's7', 's8', 's9', 's11'].forEach(id => whoosh(SCH[id].t0 + .2, .11));
// carousel: breakdown, riser, drop
riser(DROP - 1.6, DROP, .17); impact(DROP, .45);
[41, 53, 57, 60, 64].forEach((m, k) => piano(DROP + k * .02, m, .46, 1.6));

// UI sounds (moments that survive the cut)
[23.15, 56.2, 57.2, 72.9].forEach(T => click(M(T)));
click(SCH.s12.t0 + (110.25 - 106), .09);   // the cursor presses Try it now
[23.35, 73.1].forEach(T => chime(M(T)));
[30.7, 41.3, 55.2, 89.8].forEach(T => sparkle(M(T)));
sparkle(M(48.6), .035);

// word montage: a piano chord on every word, then the end card
for (let k = 0; k < 7; k++) { const t = M(101.4 + k * .6), c = CH[k % 4];   // one per montage word
  [c.lh[0], ...c.rh.map(x => x + 12)].forEach((m, j) => piano(t + j * .01, m, k === 6 ? .5 : .42, .5));
  voice(B.drum, t, .6, tt => sn(55 * tt) * Math.exp(-tt * 6) * att(tt, .004), .3); }
riser(M(104.4), END, .14); impact(END, .55);
[29, 41, 48, 57, 60, 64, 67, 72, 76].forEach((m, k) => piano(END + k * .04, m, k < 2 ? .5 : .44, 5));
[72, 76, 79, 84, 88].forEach((m, i) => bell(END + i * .04, m, .035, -.4 + i * .2, 2.4));

/* ---------- sidechain, reverb, delay, master ---------- */
const duck = new Float32Array(N).fill(1);
for (let n = 0; ; n++) { const t = G0 + n * BEAT; if (t >= MONT) break; if (!on(t, KICK)) continue;
  const i0 = Math.round(t * SR); for (let j = 0; j < .4 * SR && i0 + j < N; j++) { const tt = j / SR; const d = 1 - .4 * Math.exp(-tt / .12) * att(tt, .004); if (d < duck[i0 + j]) duck[i0 + j] = d; } }

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
const D = Math.round(.375 * SR), dl = [new Float32Array(N), new Float32Array(N)];
let dlL = 0, dlR = 0;
for (let n = 0; n < N; n++) {
  const bl = n >= D ? dl[1][n - D] : 0, br = n >= D ? dl[0][n - D] : 0;
  dlL += .35 * (bl - dlL); dlR += .35 * (br - dlR);
  dl[0][n] = B.dly[0][n] + dlL * .33; dl[1][n] = B.dly[1][n] + dlR * .33;
}

const L = new Float32Array(N), R = new Float32Array(N); let peak = 0;
for (let n = 0; n < N; n++) {
  const t = n / SR, dk = duck[n], da = .8 + .2 * dk;
  const fade = clamp(t / .03) * (t > DUR - 1.4 ? Math.max(0, (DUR - t) / 1.4) : 1);
  for (let ch = 0; ch < 2; ch++) {
    const v = B.keys[ch][n] * da + B.pad[ch][n] * dk + B.bass[ch][n] * dk + B.drum[ch][n] + B.fx[ch][n] + rev[ch][n] * 2.4 + (dl[ch][n] - B.dly[ch][n]) * .8;
    const o = Math.tanh(v * 1.05) * fade; (ch ? R : L)[n] = o; peak = Math.max(peak, Math.abs(o));
  }
}
const g = .89 / peak, buf = Buffer.alloc(44 + N * 4);
buf.write('RIFF', 0); buf.writeUInt32LE(36 + N * 4, 4); buf.write('WAVE', 8); buf.write('fmt ', 12);
buf.writeUInt32LE(16, 16); buf.writeUInt16LE(1, 20); buf.writeUInt16LE(2, 22); buf.writeUInt32LE(SR, 24);
buf.writeUInt32LE(SR * 4, 28); buf.writeUInt16LE(4, 32); buf.writeUInt16LE(16, 34); buf.write('data', 36); buf.writeUInt32LE(N * 4, 40);
for (let n = 0; n < N; n++) { buf.writeInt16LE(Math.round(clamp(L[n] * g, -1, 1) * 32767), 44 + n * 4); buf.writeInt16LE(Math.round(clamp(R[n] * g, -1, 1) * 32767), 46 + n * 4); }
fs.writeFileSync(process.argv[2] || 'music.wav', buf);
console.log('wrote', process.argv[2] || 'music.wav', 'duration', DUR.toFixed(2), 'drop', DROP.toFixed(2), 'end', END.toFixed(2));
