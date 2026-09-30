// Renders videos/<slug>/source.html to videos/<slug>/<slug>-<ratio>.mp4.
//
// The page must expose window.render(t) (draws the frame at t seconds, deterministically)
// and window.DURATION (total seconds). See videos/README.md for the full contract.
//
// Usage:
//   node render.js <slug>                   1920x1080 MP4
//   node render.js <slug> --size 1080x1920  vertical MP4
//   node render.js <slug> --stills 2,8.5,20 JPEG previews into <slug>/stills/ (gitignored)
//   node render.js <slug> --fps 60
//   node render.js <slug> --variant light   opens source.html#light, writes <slug>-light-<ratio>.mp4
//   node render.js <slug> --range 22.6,30.2  renders only that span to <slug>-<ratio>-range.mp4, for splicing a fix into an existing render

const puppeteer = require('puppeteer-core');
const ffmpeg = require('ffmpeg-static');
const { spawn } = require('child_process');
const fs = require('fs');
const path = require('path');

const CHROME = process.env.CHROME_PATH || '/usr/bin/google-chrome';
const RATIOS = { '1920x1080': '16x9', '1080x1920': '9x16', '1080x1080': '1x1', '1080x1350': '4x5' };

const args = process.argv.slice(2);
const slug = args[0];
const opt = name => { const i = args.indexOf(`--${name}`); return i === -1 ? null : args[i + 1]; };
if (!slug || slug.startsWith('--')) {
  console.error('usage: node render.js <slug> [--size WxH] [--fps N] [--stills t1,t2,...] [--variant name]');
  process.exit(1);
}

const dir = path.resolve(__dirname, '..', slug);
const source = path.join(dir, 'source.html');
if (!fs.existsSync(source)) {
  console.error(`no source.html in ${dir}`);
  process.exit(1);
}

const size = opt('size') || '1920x1080';
const [width, height] = size.split('x').map(Number);
const ratio = RATIOS[size] || size;
const fps = Number(opt('fps') || 30);
const stills = opt('stills');
const variant = opt('variant');
const tag = variant ? `${variant}-${ratio}` : ratio;
const range = opt('range') ? opt('range').split(',').map(Number) : null;

(async () => {
  const browser = await puppeteer.launch({ executablePath: CHROME, headless: 'new', args: ['--no-sandbox'] });
  const page = await browser.newPage();
  await page.setViewport({ width, height });
  await page.goto('file://' + source + (variant ? '#' + variant : ''), { waitUntil: 'networkidle0' });
  await page.evaluate(() => document.fonts.ready);

  const { duration, hasRender } = await page.evaluate(() => ({ duration: window.DURATION, hasRender: typeof window.render === 'function' }));
  if (!hasRender || !duration) {
    console.error('source.html must define window.render(t) and window.DURATION');
    process.exit(1);
  }

  if (stills) {
    const out = path.join(dir, 'stills');
    fs.mkdirSync(out, { recursive: true });
    for (const t of stills.split(',').map(Number)) {
      await page.evaluate(t => window.render(t), t);
      await page.screenshot({ path: path.join(out, `${tag}-${t}s.jpg`), type: 'jpeg', quality: 75 });
    }
    console.log(`stills written to ${out}`);
  } else {
    const file = path.join(dir, `${slug}-${tag}${range ? '-range' : ''}.mp4`);
    const ff = spawn(ffmpeg, ['-y', '-loglevel', 'error', '-f', 'image2pipe', '-framerate', String(fps), '-i', '-',
      '-c:v', 'libx264', '-pix_fmt', 'yuv420p', '-crf', '18', '-preset', 'medium', '-movflags', '+faststart', file],
      { stdio: ['pipe', 'inherit', 'inherit'] });
    const frames = Math.round(duration * fps);
    const [f0, f1] = range ? [Math.round(range[0] * fps), Math.min(frames, Math.round(range[1] * fps))] : [0, frames];
    for (let i = f0; i < f1; i++) {
      await page.evaluate(t => window.render(t), i / fps);
      const buf = await page.screenshot({ type: 'jpeg', quality: 92 });
      if (!ff.stdin.write(buf)) await new Promise(r => ff.stdin.once('drain', r));
      if (i % (fps * 5) === 0) console.log(`frame ${i}/${frames}`);
    }
    ff.stdin.end();
    await new Promise(r => ff.on('close', r));
    console.log(`wrote ${file}`);
  }
  await browser.close();
})();
