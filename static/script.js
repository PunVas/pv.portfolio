/* ──────────────────────────────────────────────
   Portfolio · script.js
   - Interactive Web Terminal (ANSI -> HTML)
   - Domain auto-detection
   - Stat counter, Scroll reveal, Contact form
────────────────────────────────────────────── */

'use strict';

// ─── Auto-detect domain & SSH command ─────────
// Try to read from the env-driven HTML first
const sshCmdEl = document.getElementById('ssh-cmd');
const SSH_CMD = sshCmdEl ? sshCmdEl.textContent : `ssh ${window.location.hostname || 'portfolio.vaswani.dev'} -p 2222`;
const HOST = window.location.hostname || 'portfolio.vaswani.dev';

// ─── ANSI to HTML ─────────────────────────────
function ansiToHtml(text) {
    if (!text) return '';
    let html = text
        .replace(/\033\[0m/g, '</span>')
        .replace(/\033\[1m/g, '<span class="t-bold">')
        .replace(/\033\[1;32m/g, '<span class="t-ok">')
        .replace(/\033\[1;96m/g, '<span class="t-sys">')
        .replace(/\033\[1;33m/g, '<span class="t-warn">')
        .replace(/\033\[1;31m/g, '<span style="color:var(--accent2)">')
        .replace(/\033\[1;35m/g, '<span class="t-sys" style="color:#d16ee8">')
        .replace(/\033\[2m/g, '<span class="t-dim">')
        .replace(/\r\n/g, '<br>')
        .replace(/\n/g, '<br>');

    // Auto-link URLs (very basic regex for https://...)
    // Not linking inside existing HTML tags is tricky, but our text doesn't have raw HTML anyway.
    html = html.replace(/(https?:\/\/[^\s<]+)/g, '<a href="$1" target="_blank" class="t-link">$1</a>');
    return html;
}

// ─── PUNEET — box-drawing ASCII art ──────────
// Using ██╗/╚═╝ box-drawing style (matches user's preferred style)
const PUNEET_BANNER = [
    ' ██████╗ ██╗   ██╗███╗  ██╗███████╗███████╗████████╗',
    ' ██╔══██╗██║   ██║████╗ ██║██╔════╝██╔════╝╚══██╔══╝',
    ' ██████╔╝██║   ██║██╔██╗██║█████╗  █████╗     ██║   ',
    ' ██╔═══╝ ██║   ██║██║╚████║██╔══╝  ██╔══╝     ██║   ',
    ' ██║     ╚██████╔╝██║ ╚███║███████╗███████╗   ██║   ',
    ' ╚═╝      ╚═════╝ ╚═╝  ╚══╝╚══════╝╚══════╝   ╚═╝   ',
];

// ─── Boot sequence (smart + a bit funny) ──────
function getBootLines() {
    return [
        { cls: 'dim', text: '' },
        { cls: 'ok', text: `[init] POST check passed. RAM: functional. Ego: in check.....    OK` },
        { cls: 'ok', text: `[init] loading 10,000 lines of Go code...[singleflight]......    OK` },
        { cls: 'sys', text: `[sys]  running whoami.......................... puneet vaswani` },
        { cls: 'sys', text: `[sys]  caffeine level.......................... CRITICALLY LOW` },
        { cls: 'warn', text: `[net]  pinging ${HOST}...` + '.'.repeat(Math.max(0, 36 - HOST.length)) + `    OK` },
        { cls: 'ok', text: `[db]   pulling work experience from Linkiss Korea..........    OK` },
        { cls: 'ok', text: `[db]   indexing 4 production projects (0 left as exercises).    OK` },
        { cls: 'ok', text: `[ai]   XLM-RoBERTa warmed up. Hindi NLI standing by.........    OK` },
        { cls: 'ok', text: `[perf] 9k req/sec backend ready. Redis singleflight armed..    OK` },
        { cls: 'ok', text: `[sec]  no hardcoded secrets found. (we checked. twice.)......    OK` },
        { cls: 'bold', text: `[boot] all systems go — type 'help' and let's talk.` },
    ];
}

function buildTerminal() {
    const body = document.getElementById('terminal-body');
    if (!body) return;

    // Set SSH command dynamically
    const sshEl = document.getElementById('ssh-cmd');
    if (sshEl) sshEl.textContent = SSH_CMD;

    body.innerHTML = '';
    const lines = getBootLines();
    let i = 0;

    function addLine() {
        if (i >= lines.length) { addBanner(); return; }
        const { cls, text } = lines[i++];
        const span = document.createElement('span');
        span.className = `t-line t-${cls}`;
        span.textContent = text;
        body.appendChild(span);
        body.scrollTop = body.scrollHeight;
        setTimeout(addLine, 65);
    }

    function addBanner() {
        body.appendChild(Object.assign(document.createElement('span'), { className: 't-line' }));

        PUNEET_BANNER.forEach((row, idx) => {
            setTimeout(() => {
                const s = document.createElement('span');
                s.className = 't-line t-sys';
                s.style.fontWeight = '700';
                s.textContent = row;
                body.appendChild(s);
                body.scrollTop = body.scrollHeight;
                if (idx === PUNEET_BANNER.length - 1) addPrompt();
            }, idx * 50);
        });
    }

    let isTyping = false;
    let currentInput = '';
    let cursorSpan;

    function handleKeydown(e) {
        if (!isTyping) return;
        // Don't intercept if user is typing in contact form
        if (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA') return;

        if (e.key === 'Enter') {
            e.preventDefault();
            processCommand(currentInput);
        } else if (e.key === 'Backspace') {
            currentInput = currentInput.slice(0, -1);
            updateInputDisplay();
        } else if (e.key.length === 1 && !e.ctrlKey && !e.metaKey) {
            currentInput += e.key;
            updateInputDisplay();
        }
    }
    // Attach listener to window so we grab keys anywhere
    window.addEventListener('keydown', handleKeydown);

    function updateInputDisplay() {
        if (!cursorSpan) return;
        const textSpan = cursorSpan.previousElementSibling;
        if (textSpan && textSpan.className === 't-input') {
            textSpan.textContent = currentInput;
        }
        body.scrollTop = body.scrollHeight;
    }

    async function processCommand(cmd) {
        if (cursorSpan) cursorSpan.remove(); // lock input visually
        isTyping = false;

        const cleanCmd = cmd.trim();
        if (!cleanCmd) {
            currentInput = '';
            addPrompt(false);
            return;
        }

        if (cleanCmd === 'clear') {
            body.innerHTML = '';
            currentInput = '';
            addPrompt(false);
            return;
        }

        try {
            const res = await fetch('/api/cmd', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ command: cleanCmd })
            });

            if (cleanCmd === 'exit' || cleanCmd === 'contact') {
                const term = document.querySelector('.terminal-window');
                if (term.classList.contains('fullscreen')) {
                    toggleFullscreen();
                }

                if (cleanCmd === 'contact') {
                    const contactSection = document.getElementById('contact');
                    if (contactSection) {
                        contactSection.scrollIntoView({ behavior: 'smooth' });
                    }
                    const contactSpan = document.createElement('span');
                    contactSpan.className = 't-line t-ok';
                    contactSpan.innerHTML = '<br>Redirecting to contact form...<br><br>';
                    body.appendChild(contactSpan);
                } else {
                    const exitSpan = document.createElement('span');
                    exitSpan.className = 't-line t-sys';
                    exitSpan.innerHTML = '<br>Session ended. Returning to standard view...<br><br>';
                    body.appendChild(exitSpan);
                }

                currentInput = '';
                addPrompt(false);
                return;
            }

            const data = await res.json();

            const outSpan = document.createElement('span');
            outSpan.className = 't-line';
            outSpan.innerHTML = ansiToHtml(data.output);
            body.appendChild(outSpan);
        } catch (e) {
            const errSpan = document.createElement('span');
            errSpan.className = 't-line t-warn';
            errSpan.innerHTML = '<br>Network error executing command.<br><br>';
            body.appendChild(errSpan);
        }

        currentInput = '';
        addPrompt(false);
    }

    let firstPrompt = true;
    function addPrompt(showHeader = true) {
        setTimeout(() => {
            if (firstPrompt && showHeader) {
                [{ cls: 't-dim', text: `  Systems-focused Backend Developer  ·  IIIT Bhopal  ·  India` },
                { cls: 't-dim', text: `  ${'─'.repeat(60)}` }
                ].forEach(({ cls, text }) => {
                    const s = document.createElement('span');
                    s.className = `t-line ${cls}`;
                    s.textContent = text;
                    body.appendChild(s);
                });
                firstPrompt = false;
            }

            const prompt = document.createElement('span');
            prompt.className = 't-line';
            prompt.innerHTML =
                '<span class="t-ok">puneet@portfolio</span>' +
                '<span class="t-dim">:</span>' +
                '<span class="t-sys">~</span>' +
                '<span class="t-dim">$ </span>' +
                '<span class="t-input"></span>' +
                '<span class="cursor"></span>';
            body.appendChild(prompt);

            cursorSpan = prompt.querySelector('.cursor');
            isTyping = true;
            body.scrollTop = body.scrollHeight;
        }, showHeader ? 250 : 0);
    }

    setTimeout(addLine, 400);
}

// ─── Stat counter ──────────────────────────────
function animateCounters() {
    document.querySelectorAll('.stat-num[data-target]').forEach(el => {
        if (el.classList.contains('animated')) return;
        el.classList.add('animated');

        const target = parseFloat(el.dataset.target);
        if (isNaN(target)) return;

        const originalText = el.textContent;
        const suffix = originalText.replace(/[0-9.]/g, '');

        let current = 0;
        const duration = 1500;
        const start = performance.now();

        function update(now) {
            const elapsed = now - start;
            const progress = Math.min(elapsed / duration, 1);

            // Ease out cubic
            const ease = 1 - Math.pow(1 - progress, 3);
            current = target * ease;

            const display = target % 1 === 0 ? Math.floor(current) : current.toFixed(2);
            el.textContent = display + suffix;

            if (progress < 1) {
                requestAnimationFrame(update);
            } else {
                el.textContent = originalText; // Ensure exact match at end
            }
        }
        requestAnimationFrame(update);
    });
}

// ─── Scroll reveal ─────────────────────────────
function initReveal() {
    const obs = new IntersectionObserver((entries) => {
        entries.forEach(e => {
            // Toggles visible class based on intersection (Apple-style bidirectional)
            if (e.isIntersecting) {
                e.target.classList.add('visible');
                if (e.target.closest('#stats')) animateCounters();
            } else {
                e.target.classList.remove('visible');
            }
        });
    }, {
        threshold: 0.1,
        rootMargin: '0px 0px -50px 0px' // Trigger slightly before it hits bottom
    });
    document.querySelectorAll('.reveal').forEach(el => obs.observe(el));
}

// ─── Copy SSH command ──────────────────────────
function copySSH() {
    const cmd = document.getElementById('ssh-cmd').textContent;
    navigator.clipboard.writeText(cmd).then(() => {
        const btn = document.getElementById('copy-ssh');
        btn.textContent = 'copied!';
        setTimeout(() => { btn.textContent = 'copy'; }, 2000);
    });
}
window.copySSH = copySSH;

// ─── Contact form AJAX ─────────────────────────
async function submitContact(e) {
    e.preventDefault();
    const btn = document.getElementById('submit-btn');
    const status = document.getElementById('form-status');
    const name = document.getElementById('name').value.trim();
    const msg = document.getElementById('message').value.trim();

    btn.disabled = true;
    btn.textContent = '⏳ Sending...';
    status.textContent = '';
    status.className = 'form-status';

    try {
        const res = await fetch('/api/contact', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name, message: msg }),
        });
        const data = await res.json();
        if (res.ok && data.ok) {
            status.textContent = '✓ ' + data.message;
            status.className = 'form-status success';
            document.getElementById('contact-form').reset();
        } else {
            status.textContent = '✗ ' + (data.message || 'Something went wrong.');
            status.className = 'form-status error';
        }
    } catch {
        status.textContent = '✗ Network error. Try again.';
        status.className = 'form-status error';
    } finally {
        btn.disabled = false;
        btn.textContent = '✓ Send via Webhook';
    }
}
window.submitContact = submitContact;

// ─── Scroll-driven Sections (Apple-style) ──────
function initScrollSections() {
    const sections = document.querySelectorAll('.scroll-section');

    window.addEventListener('scroll', () => {
        sections.forEach(section => {
            const wrapper = section.querySelector('.sticky-wrapper');
            const items = section.querySelectorAll('.scroll-item');
            if (items.length === 0) return;

            const rect = section.getBoundingClientRect();
            const sectionTop = section.offsetTop;
            const sectionHeight = section.offsetHeight;
            const viewportHeight = window.innerHeight;

            // Calculate progress: 0 when top enters, 1 when bottom leaves
            let progress = (window.scrollY - sectionTop) / (sectionHeight - viewportHeight);
            progress = Math.max(0, Math.min(1, progress));

            // Divide progress by number of items
            const itemProgress = progress * items.length;

            items.forEach((item, index) => {
                const threshold = index;
                const nextThreshold = index + 1;

                // Show item if progress is within its range
                if (itemProgress >= threshold - 0.2 && itemProgress < nextThreshold) {
                    item.classList.add('active');
                    item.classList.remove('exit');
                } else if (itemProgress >= nextThreshold) {
                    item.classList.remove('active');
                    item.classList.add('exit');
                } else {
                    item.classList.remove('active', 'exit');
                }
            });
        });
    });
}

// ─── Terminal Fullscreen Toggle ────────────────
function toggleFullscreen() {
    const term = document.querySelector('.terminal-window');
    const maxIcon = document.getElementById('maximize-icon');
    const minIcon = document.getElementById('minimize-icon');
    const body = document.getElementById('terminal-body');

    term.classList.toggle('fullscreen');

    if (term.classList.contains('fullscreen')) {
        maxIcon.style.display = 'none';
        minIcon.style.display = 'inline-block';
        // Remove tilt when fullscreen
        term.vanillaTilt && term.vanillaTilt.destroy();
    } else {
        maxIcon.style.display = 'inline-block';
        minIcon.style.display = 'none';
        // Re-add tilt
        VanillaTilt.init(term);
    }

    // Scroll to bottom after resize
    setTimeout(() => {
        body.scrollTop = body.scrollHeight;
    }, 300);
}
window.toggleFullscreen = toggleFullscreen;

// ─── Init ──────────────────────────────────────
document.addEventListener('DOMContentLoaded', () => {
    buildTerminal();
    initScrollSections();

    // Auto-fullscreen on terminal interaction
    const term = document.querySelector('.terminal-window');
    const toggleBtn = document.getElementById('term-toggle');

    // Stop propagation on the button so it doesn't trigger the window's click listener
    toggleBtn.addEventListener('click', (e) => e.stopPropagation());

    term.addEventListener('click', () => {
        if (!term.classList.contains('fullscreen')) {
            toggleFullscreen();
        }
    });

    // Handle keydown to also trigger fullscreen
    const originalHandleKeydown = window.handleKeydown;
    window.addEventListener('keydown', (e) => {
        // If it's a typing key and terminal isn't fullscreen, expand it
        if (!term.classList.contains('fullscreen') && e.key.length === 1 && !e.ctrlKey && !e.metaKey) {
            // Check if we're not in an input/textarea
            if (e.target.tagName !== 'INPUT' && e.target.tagName !== 'TEXTAREA') {
                toggleFullscreen();
            }
        }
    });

    // Check if initReveal or handleReveal exists (subagent might have changed it)
    if (typeof initReveal === 'function') initReveal();
    else if (typeof handleReveal === 'function') handleReveal();

    // Scroll listener for hero hint and page state
    window.addEventListener('scroll', () => {
        if (window.scrollY > 100) {
            document.body.classList.add('scrolled');
        } else {
            document.body.classList.remove('scrolled');
        }
    });

    // Re-trigger reveal on scroll
    window.addEventListener('scroll', () => {
        if (typeof handleReveal === 'function') handleReveal();
        else if (typeof initReveal === 'function') initReveal();
    });
});
