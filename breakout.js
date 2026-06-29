// breakout.js
#!/usr/bin/env node
'use strict';

const blessed = require('blessed');
const fs = require('fs');
const path = require('path');
const os = require('os');

const RECORD_FILE = path.join(os.homedir(), '.breakout_record.json');
function loadRecord() {
    try { return JSON.parse(fs.readFileSync(RECORD_FILE)).record || 0; } catch { return 0; }
}
function saveRecord(record) {
    fs.writeFileSync(RECORD_FILE, JSON.stringify({ record }));
}

let speed = 50;
if (process.argv.includes('-s') && process.argv.length > process.argv.indexOf('-s')+1) {
    speed = parseInt(process.argv[process.argv.indexOf('-s')+1]) || 50;
}

const screen = blessed.screen({
    smartCSR: true,
    title: 'Breakout',
    fullUnicode: true,
});
const height = screen.height;
const width = screen.width;
if (height < 25 || width < 50) {
    console.log('Terminal too small (min 25x50)');
    process.exit(1);
}

const H = height - 2;
const W = width - 2;
const padY = 1, padX = 1;

let paddleW = 8;
let paddleX = Math.floor((W - paddleW) / 2);
const paddleY = H - 2;
let ballX = paddleX + Math.floor(paddleW/2);
let ballY = paddleY - 1;
let ballDX = 1, ballDY = -1;
let ballSpeed = 1.5;
let ballActive = false;
let gameOver = false;
let lives = 3, score = 0, level = 1;
let best = loadRecord();
let frameTime = speed;

let blocks = [];
let bonuses = [];

function generateLevel() {
    blocks = [];
    const rows = 5 + Math.floor(level / 2);
    const cols = 8 + level;
    for (let r=0; r<rows; r++) {
        for (let c=0; c<cols; c++) {
            if (c*2+2 < W-2) {
                let hp = 1;
                const rnd = Math.random()*100;
                if (rnd < 70) hp = 1;
                else if (rnd < 90) hp = 2;
                else hp = 3;
                const color = Math.min(hp+1, 7);
                blocks.push({x: c*2+2, y: r*2+2, hp, color});
            }
        }
    }
}
generateLevel();

function draw() {
    screen.clear();
    // Рамка
    for (let y=0; y<=H; y++) {
        screen.fillRegion('|', padX-1, padY+y, padX, padY+y+1, blessed.colors.white, blessed.colors.black);
        screen.fillRegion('|', padX+W, padY+y, padX+W+1, padY+y+1, blessed.colors.white, blessed.colors.black);
    }
    for (let x=0; x<W+2; x++) {
        screen.fillRegion('-', padX+x-1, padY-1, padX+x, padY, blessed.colors.white, blessed.colors.black);
    }
    // Платформа
    for (let i=0; i<paddleW; i++) {
        screen.fillRegion('=', padX+paddleX+i, padY+paddleY, padX+paddleX+i+1, padY+paddleY+1, blessed.colors.yellow, blessed.colors.black);
    }
    // Мяч
    if (ballActive || !gameOver) {
        screen.fillRegion('O', padX+Math.floor(ballX), padY+Math.floor(ballY), padX+Math.floor(ballX)+1, padY+Math.floor(ballY)+1, blessed.colors.cyan, blessed.colors.black);
    }
    // Блоки
    for (const b of blocks) {
        const color = b.color <= 7 ? blessed.colors.cyan : blessed.colors.white;
        for (let i=0; i<2; i++) {
            for (let j=0; j<2; j++) {
                screen.fillRegion('#', padX+b.x+i, padY+b.y+j, padX+b.x+i+1, padY+b.y+j+1, color, blessed.colors.black);
            }
        }
    }
    // Бонусы
    for (const bon of bonuses) {
        const sym = ['W','S','L','F'][bon.type];
        screen.fillRegion(sym, padX+bon.x, padY+bon.y, padX+bon.x+1, padY+bon.y+1, blessed.colors.magenta, blessed.colors.black);
    }
    // Счёт
    screen.setContent(0, 2, `Score: ${score}`, blessed.colors.white);
    screen.setContent(0, Math.floor(W/2)-4, `Best: ${best}`, blessed.colors.white);
    screen.setContent(0, W-20, `Lives: ${lives}  Level: ${level}`, blessed.colors.white);
    if (gameOver) {
        screen.setContent(Math.floor(H/2), Math.floor((W - 30)/2), 'GAME OVER! Press R to restart, Q to quit', blessed.colors.red);
    }
    screen.render();
}

function update() {
    if (gameOver) { draw(); return; }

    if (ballActive) {
        ballX += ballDX * ballSpeed;
        ballY += ballDY * ballSpeed;
        if (ballX <= 0 || ballX >= W-1) ballDX *= -1;
        if (ballY <= 0) ballDY *= -1;
        if (Math.floor(ballY) === paddleY-1 && paddleX <= Math.floor(ballX) && Math.floor(ballX) < paddleX+paddleW) {
            ballDY *= -1;
            const offset = (ballX - paddleX) / paddleW;
            ballDX = (offset - 0.5) * 2;
            if (Math.abs(ballDX) < 0.3) ballDX = ballDX >= 0 ? 0.5 : -0.5;
            process.stdout.write('\x07');
        }
        for (let i=0; i<blocks.length; i++) {
            const b = blocks[i];
            if (b.x <= ballX && ballX < b.x+2 && b.y <= ballY && ballY < b.y+2) {
                b.hp--;
                if (b.hp <= 0) {
                    score += 10;
                    if (Math.random() < 0.15) {
                        const type = Math.floor(Math.random()*4);
                        bonuses.push({x: b.x, y: b.y, type});
                    }
                    blocks.splice(i,1);
                    process.stdout.write('\x07');
                }
                ballDY *= -1;
                break;
            }
        }
        if (ballY >= H) {
            lives--;
            if (lives <= 0) {
                gameOver = true;
                if (score > best) { best = score; saveRecord(best); }
            } else {
                ballX = paddleX + Math.floor(paddleW/2);
                ballY = paddleY - 1;
                ballDX = 1; ballDY = -1;
                ballActive = false;
            }
        }
        if (blocks.length === 0) {
            level++;
            ballSpeed += 0.3;
            generateLevel();
            ballX = paddleX + Math.floor(paddleW/2);
            ballY = paddleY - 1;
            ballDX = 1; ballDY = -1;
            ballActive = false;
        }
    }

    for (let i=0; i<bonuses.length; i++) {
        const bon = bonuses[i];
        bon.y += 1;
        if (bon.y >= H) { bonuses.splice(i,1); i--; continue; }
        if (bon.y === paddleY && paddleX <= bon.x && bon.x < paddleX+paddleW) {
            if (bon.type === 0) paddleW = Math.min(16, paddleW+4);
            else if (bon.type === 1) ballSpeed += 0.5;
            else if (bon.type === 2) lives++;
            else if (bon.type === 3) {
                for (const b of blocks) b.hp -= 1;
            }
            bonuses.splice(i,1);
            i--;
            process.stdout.write('\x07');
        }
    }
    draw();
    setTimeout(update, frameTime);
}

screen.key(['left','a'], function() { paddleX = Math.max(0, paddleX-2); });
screen.key(['right','d'], function() { paddleX = Math.min(W-paddleW, paddleX+2); });
screen.key(['space'], function() {
    if (!ballActive && !gameOver) ballActive = true;
});
screen.key(['r','R'], function() {
    if (gameOver) {
        paddleX = Math.floor((W - paddleW) / 2);
        ballX = paddleX + Math.floor(paddleW/2);
        ballY = paddleY - 1;
        ballDX = 1; ballDY = -1;
        ballActive = false;
        lives = 3; score = 0; level = 1;
        gameOver = false;
        bonuses = [];
        paddleW = 8;
        ballSpeed = 1.5;
        generateLevel();
        draw();
    }
});
screen.key(['q','Q'], function() { process.exit(0); });

draw();
update();
