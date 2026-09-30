// Original score for ai-chat-promo, synthesized from scratch (no samples, no licensed audio).
// Every hit is placed on a timestamp from source.html, so the music follows the edit.
//
// Usage: node music.js <out.wav>
// Then:  ffmpeg -i out.wav -af loudnorm=I=-14:TP=-1.5 -b:a 192k ai-chat-promo-music.mp3

const fs = require('fs');
const SR = 44100, DUR = 113, N = SR * DUR;
const mk = () => [new Float32Array(N), new Float32Array(N)];
const B = { pad: mk(), arp: mk(), drum: mk(), bass: mk(), fx: mk(), rev: mk(), dly: mk() };

const TS = 8192, ST = new Float32Array(TS);
for (let i = 0; i < TS; i++) ST[i] = Math.sin(2 * Math.PI * i / TS);
const sn = ph => ST[((ph - Math.floor(ph)) * TS) | 0];
const mtof = m => 440 * Math.pow(2, (m - 69) / 12);
let seed = 7; const rnd = () => { seed = (seed * 16807) % 2147483647; return seed / 2147483647 * 2 - 1; };
const clamp = (x, a = 0, b = 1) => Math.max(a, Math.min(b, x));
const on = (t, ranges) => ranges.some(([a, b]) => t >= a && t < b);

/* ---------- arrangement ---------- */
const BEAT = 0.5, G0 = 6.6;                           // 120 BPM grid, starts when the chat is revealed
const snap = x => G0 + Math.round((x - G0) / BEAT) * BEAT;
const DROP = snap(58.2);                              // carousel templates fan out
const KICK = [[snap(13.3), 51.6], [DROP, snap(101.2)]];
const HAT = [[snap(10.6), 51.6], [DROP, snap(101.2)]];
const HAT16 = [[snap(25.8), 51.6], [snap(64.2), snap(101.2)]];
const CLAP = [[snap(25.8), 51.6], [DROP, snap(101.2)]];
const LEAD = [[DROP, snap(64.2)], [snap(85), snap(101.2)]];

// Fmaj7, Am7, Dm7, Bbmaj7 (two bars each), bass roots
const CH = [[53, 57, 60, 64], [57, 60, 64, 67], [50, 53, 57, 60], [46, 50, 53, 57]];
const BS = [41, 45, 38, 34];
const chordAt = t => (t < G0 ? 0 : Math.floor((t - G0) / 4) % 4);

// brightness curve: how open the "filter" is (number of audible harmonics)
function bright(t) {
  if (t < 3.2) return 1.2 + t * .3;
  if (t < 6.6) return 2.4 + (t - 3.2) * .3;
  if (t < 13.1) return 2.6;
  if (t < 51.6) return 3.6;
  if (t < DROP) return 1.6 + 3.4 * Math.pow((t - 51.6) / (DROP - 51.6), 2);  // build into the drop
  if (t < 101.2) return 4.2;
  if (t < 106) return 3.2;
  return 6 - (t - 106) * .5;
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
const att = (tt, a) => (tt < a ? tt / a : 1);

/* ---------- instruments ---------- */
function pluck(m, br) {
  const f = mtof(m), H = Math.min(10, Math.floor(11000 / f));
  return tt => { let s = 0; for (let h = 1; h <= H; h++) s += sn(f * h * tt) / h * Math.exp(-tt * (5 + h * h * .5 / br)); return s * att(tt, .002); };
}
function bell(t, m, gain, pan = 0, decay = .9, rs = .45, ds = .2) {
  const f = mtof(m);
  voice(B.fx, t, decay * 5, tt => (sn(f * tt) + .4 * sn(2 * f * tt) * Math.exp(-tt * 3) + .18 * sn(3.01 * f * tt) * Math.exp(-tt * 6)) * Math.exp(-tt / decay) * att(tt, .003), gain, pan, rs, ds);
}
function kick(t) {
  voice(B.drum, t, .5, tt => { const ph = 45 * tt + 100 * (1 - Math.exp(-30 * tt)) / 30; return sn(ph) * Math.exp(-tt * 7) * att(tt, .002) + rnd() * Math.exp(-tt * 400) * .25; }, .62);
}
function hat(t, g, dec, pan) {
  let p = 0; voice(B.drum, t, .12, tt => { const x = rnd(), y = x - p; p = x; return y * Math.exp(-tt * dec); }, g, pan, .08);
}
function clap(t) {
  let lp = 0, lp2 = 0;
  voice(B.drum, t, .3, tt => {
    const env = (tt < .01 ? 1 : tt < .02 ? .7 : 1) * Math.exp(-tt * 18) + (tt > .02 ? Math.exp(-(tt - .02) * 14) * .6 : 0);
    const x = rnd(); lp += .35 * (x - lp); lp2 += .08 * (lp - lp2); return (lp - lp2) * env * 2.2;
  }, .16, -.1, .35);
}
function bassNote(t, m) {
  const f = mtof(m);
  voice(B.bass, t, .26, tt => (sn(f * tt) + .3 * sn(2 * f * tt) + .12 * sn(3 * f * tt)) * att(tt, .008) * Math.exp(-tt * 5) * (tt > .22 ? (.26 - tt) / .04 : 1), .3);
}
function svfNoise(fcFn, ampFn) { // noise through a swept state-variable bandpass
  let low = 0, band = 0;
  return (tt) => { const fc = Math.min(7000, fcFn(tt)), f = 2 * Math.sin(Math.PI * fc / SR); const x = rnd(); low += f * band; const high = x - low - .5 * band; band += f * high; return band * ampFn(tt); };
}
function whoosh(tc, gain = .14, dur = .9) {
  const t0 = tc - dur * .7;
  voice(B.fx, t0, dur, svfNoise(tt => 300 * Math.pow(5000 / 300, tt / dur), tt => { const x = tt / dur; return Math.pow(Math.sin(Math.PI * Math.min(1, x / .7) / 2), 2) * (x > .7 ? Math.exp(-(x - .7) * 12) : 1); }), gain, 0, .3);
}
function riser(t0, t1, gain = .16, root = 65) {
  const d = t1 - t0, f0 = mtof(root);
  voice(B.fx, t0, d, svfNoise(tt => 200 * Math.pow(35, tt / d), tt => Math.pow(tt / d, 2) * (tt > d - .01 ? (d - tt) / .01 : 1)), gain, 0, .25);
  voice(B.fx, t0, d, tt => { const k = tt / d, ph = f0 * (d / Math.LN2) * (Math.pow(2, k) - 1); let s = 0; for (let h = 1; h <= 5; h++) s += sn(ph * h) / h; return s * k * k * (tt > d - .01 ? (d - tt) / .01 : 1); }, gain * .35, 0, .3);
}
function impact(t, gain = .5) {
  let lp = 0;
  voice(B.fx, t, 3.2, tt => { const ph = 28 * tt + 42 * (1 - Math.exp(-4 * tt)) / 4; const x = rnd(); lp += .05 * (x - lp); return sn(ph) * Math.exp(-tt * 1.6) * att(tt, .003) + lp * 3 * Math.exp(-tt * 7); }, gain, 0, .4);
}
function click(t, g = .07) { voice(B.fx, t, .03, tt => sn(1900 * tt) * Math.exp(-tt * 250) + rnd() * Math.exp(-tt * 700) * .3, g, .15, .05); }
function sendBlip(t) { bell(t, 88, .045, .1, .1, .2, 0); bell(t + .05, 93, .045, .1, .12, .2, 0); }
function chime(t) { [84, 88, 91].forEach((m, i) => bell(t + i * .09, m, .09, -.2 + i * .2, .7)); }
function sparkle(t, g = .045) { [84, 86, 89, 91, 93, 96].forEach((m, i) => bell(t + i * .055, m, g, -.4 + i * .16, .5)); }
function stab(t, chord, g = .1) {
  chord.forEach((m, i) => { voice(B.arp, t, 1.2, pluck(m + 12, 7), g, -.3 + i * .2, .3, .15); voice(B.arp, t, 1.2, pluck(m, 5), g * .7, .3 - i * .2, .2); });
  voice(B.drum, t, .6, tt => sn(55 * tt) * Math.exp(-tt * 6) * att(tt, .004), .35);
}

/* ---------- pad (whole timeline, chord segments crossfade) ---------- */
function pad(t0, t1, notes, gain = .05) {
  const dets = [-8, 0, 8], rel = .9, H = 8, W = new Float32Array(H + 1);
  let wb = -1;
  voice(B.pad, t0, t1 - t0 + rel, (tt, ta) => {
    const blk = (ta * SR / 512) | 0;
    if (blk !== wb) { wb = blk; const b = bright(ta); for (let h = 1; h <= H; h++) W[h] = Math.exp(-h / b) / h; }
    let s = 0;
    for (const m of notes) { const f = mtof(m); for (const d of dets) { const fd = f * Math.pow(2, d / 1200); for (let h = 1; h <= H; h++) s += sn(fd * h * tt + d * .01) * W[h]; } }
    const env = clamp(tt / .7) * (tt > t1 - t0 ? Math.max(0, 1 - (tt - (t1 - t0)) / rel) : 1);
    return s * env;
  }, gain, 0, .5);
}

/* ---------- write the score ---------- */
// pad
pad(0, G0, [53, 57, 60, 64, 67], .045);
for (let t = G0; t < 106; t += 4) pad(t, Math.min(t + 4, 106), CH[chordAt(t + .01)], .05);
pad(106, 113, [41, 53, 57, 60, 64, 67, 72], .06);

// arp (16ths) from the chat reveal until the word montage
const PAT = [0, 2, 4, 6, 7, 5, 3, 1, 0, 2, 4, 6, 7, 6, 4, 2];
for (let s = 0; ; s++) {
  const t = G0 + s * .125; if (t >= snap(101.2)) break;
  const c = CH[chordAt(t)], tones = [...c, ...c.map(x => x + 12)];
  let lvl = t < 13.1 ? .55 * clamp((t - G0) / 2) : (t >= 51.6 && t < DROP ? .6 : 1);
  const acc = s % 4 === 0 ? 1 : .7;
  voice(B.arp, t, .6, pluck(tones[PAT[s % 16]] + 12, bright(t) * 1.3), .085 * lvl * acc, s % 2 ? .3 : -.3, .18, .28);
}

// drums + bass on the grid
for (let n = 0; ; n++) {
  const t = G0 + n * BEAT; if (t >= 106) break;
  if (on(t, KICK)) { kick(t); if ((n % 4) === 1 || (n % 4) === 3) if (on(t, CLAP)) clap(t); }
  const off = t + .25;
  if (on(off, HAT)) hat(off, .075, 55, .25);
  if (on(t, HAT16)) { hat(t + .125, .03, 90, -.2); hat(t + .375, .03, 90, -.2); }
  if (on(off, KICK)) bassNote(off, BS[chordAt(off)]);
}
// lead bells over the drop and the skills/more section
for (let t = G0; t < snap(101.2); t += 1) if (on(t, LEAD)) { const c = CH[chordAt(t)]; bell(t, c[((t - G0) % 2) < .5 ? 3 : 2] + 24, .05, ((t | 0) % 2) ? .35 : -.35, 1.1, .5, .3); }

// intro: a note per word, boom on "Now just ask.", riser into the reveal
[77, 81, 84, 88].forEach((m, k) => bell(.35 + .68 * k, m, .1, -.3 + k * .2, .9));
impact(3.2, .45); [72, 76, 79, 84].forEach((m, i) => bell(3.2 + i * .03, m, .05, -.3 + i * .2, 1.6));
riser(4.1, 5.25, .14); whoosh(5.7, .18, 1.2);

// scene cuts
[13.3, 25.8, 34.8, 42.6, 51.6, 64.2, 75, 85, 92.8].forEach(t => whoosh(t, .12));
// carousel build and drop
riser(DROP - 2, DROP, .18); impact(DROP, .5);

// UI: clicks, sends, success chimes, reveal sparkles
[10.75, 11.7, 23.15, 56.2, 57.2, 72.9, 96.6].forEach(t => click(t));
[14.1, 20.4, 28.8, 40.2, 46.4, 65.7, 69.8, 76.7, 89.3].forEach(sendBlip);
[23.35, 73.1, 95.8].forEach(chime);
[30.7, 41.3, 55.2, 89.8].forEach(t => sparkle(t));
sparkle(48.6, .035);

// word montage: a stab on every word, then the end card
for (let k = 0; k < 8; k++) stab(101.4 + k * .6, CH[k % 4], k === 7 ? .12 : .09);
riser(104.4, 106, .15); impact(106, .6);
[72, 76, 79, 84, 88].forEach((m, i) => bell(106 + i * .04, m, .05, -.4 + i * .2, 2.4));

/* ---------- sidechain, fx buses, master ---------- */
const duck = new Float32Array(N).fill(1);
for (let n = 0; ; n++) { const t = G0 + n * BEAT; if (t >= 106) break; if (!on(t, KICK)) continue;
  const i0 = Math.round(t * SR); for (let j = 0; j < .4 * SR && i0 + j < N; j++) { const tt = j / SR; const d = 1 - .55 * Math.exp(-tt / .12) * att(tt, .004); if (d < duck[i0 + j]) duck[i0 + j] = d; } }

function freeverb(inL, inR) {
  const C = [1116, 1188, 1277, 1356, 1422, 1491, 1557, 1617], A = [556, 441, 341, 225], fb = .86, damp = .3;
  const out = [new Float32Array(N), new Float32Array(N)];
  [inL, inR].forEach((inp, ch) => {
    const sp = ch * 23, combs = C.map(c => ({ b: new Float32Array(c + sp), i: 0, fs: 0 })), aps = A.map(a => ({ b: new Float32Array(a + sp), i: 0 }));
    for (let n = 0; n < N; n++) {
      const x = inp[n] * .015; let o = 0;
      for (const c of combs) { const y = c.b[c.i]; c.fs = y * (1 - damp) + c.fs * damp; c.b[c.i] = x + c.fs * fb; if (++c.i >= c.b.length) c.i = 0; o += y; }
      for (const a of aps) { const bo = a.b[a.i]; const y = -o + bo; a.b[a.i] = o + bo * .5; if (++a.i >= a.b.length) a.i = 0; o = y; }
      out[ch][n] = o;
    }
  });
  return out;
}
const rev = freeverb(B.rev[0], B.rev[1]);
const D = Math.round(.375 * SR), dl = [new Float32Array(N), new Float32Array(N)];
let lpL = 0, lpR = 0;
for (let n = 0; n < N; n++) {
  const bl = n >= D ? dl[1][n - D] : 0, br = n >= D ? dl[0][n - D] : 0;
  lpL += .4 * (bl - lpL); lpR += .4 * (br - lpR);
  dl[0][n] = B.dly[0][n] + lpL * .38; dl[1][n] = B.dly[1][n] + lpR * .38;
}

const L = new Float32Array(N), R = new Float32Array(N);
let peak = 0;
for (let n = 0; n < N; n++) {
  const t = n / SR, dk = duck[n], da = .75 + .25 * dk;
  const fade = clamp(t / .05) * (t > 111.2 ? Math.max(0, (113 - t) / 1.8) : 1);
  for (let ch = 0; ch < 2; ch++) {
    let v = B.pad[ch][n] * dk + B.bass[ch][n] * dk + B.arp[ch][n] * da + B.drum[ch][n] + B.fx[ch][n] + rev[ch][n] * 2.2 + (dl[ch][n] - B.dly[ch][n]) * .9;
    v = Math.tanh(v * 1.1) * fade;
    (ch ? R : L)[n] = v; peak = Math.max(peak, Math.abs(v));
  }
}

const g = .89 / peak, buf = Buffer.alloc(44 + N * 4);
buf.write('RIFF', 0); buf.writeUInt32LE(36 + N * 4, 4); buf.write('WAVE', 8); buf.write('fmt ', 12);
buf.writeUInt32LE(16, 16); buf.writeUInt16LE(1, 20); buf.writeUInt16LE(2, 22); buf.writeUInt32LE(SR, 24);
buf.writeUInt32LE(SR * 4, 28); buf.writeUInt16LE(4, 32); buf.writeUInt16LE(16, 34); buf.write('data', 36); buf.writeUInt32LE(N * 4, 40);
for (let n = 0; n < N; n++) { buf.writeInt16LE(Math.round(clamp(L[n] * g, -1, 1) * 32767), 44 + n * 4); buf.writeInt16LE(Math.round(clamp(R[n] * g, -1, 1) * 32767), 46 + n * 4); }
fs.writeFileSync(process.argv[2] || 'music.wav', buf);
console.log('wrote', process.argv[2] || 'music.wav', 'peak', peak.toFixed(3));
