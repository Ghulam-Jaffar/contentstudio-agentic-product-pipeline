// One beat grid for picture and score (source.html loads it as a script, music.js requires it).
// 68.57 BPM: a beat is 0.875 s, a bar 3.5 s. Intro 6 beats, each capability 2 bars, end card 2 bars.
const TIMING = (() => {
  const BEAT = 0.875, INTRO = 6 * BEAT, SL = 8 * BEAT, OUTRO = 8 * BEAT;
  const OUT = INTRO + 8 * SL, DUR = OUT + OUTRO;
  return { BEAT, INTRO, SL, OUT, DUR, HEADLINE: 6 * BEAT, S: i => INTRO + i * SL };   // headline lands on beat 6 of each section
})();
if (typeof module !== 'undefined') module.exports = TIMING;
