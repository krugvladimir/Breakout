# breakout.py
#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import sys, os, random, json, time, argparse, curses
from pathlib import Path

RECORD_FILE = Path.home() / '.breakout_record.json'

def load_record():
    try:
        with open(RECORD_FILE) as f:
            return json.load(f).get('record', 0)
    except:
        return 0

def save_record(record):
    with open(RECORD_FILE, 'w') as f:
        json.dump({'record': record}, f)

def main(stdscr, speed):
    curses.curs_set(0)
    stdscr.nodelay(1)
    stdscr.timeout(0)
    curses.start_color()
    curses.use_default_colors()
    for i in range(1, 8):
        curses.init_pair(i, i, -1)

    height, width = stdscr.getmaxyx()
    if height < 25 or width < 50:
        print("Терминал слишком мал (нужно 25x50)")
        return

    H, W = height-2, width-2
    pad_y, pad_x = 1, 1

    # Игровые параметры
    paddle_w = 8
    paddle_x = (W - paddle_w) // 2
    paddle_y = H - 2
    ball_x, ball_y = paddle_x + paddle_w//2, paddle_y - 1
    ball_dx, ball_dy = 1, -1
    ball_speed = 1.5
    ball_active = False
    lives = 3
    score = 0
    level = 1
    best = load_record()
    game_over = False
    frame_time = speed / 1000.0

    # Блоки
    class Block:
        def __init__(self, x, y, hp, color):
            self.x = x; self.y = y; self.hp = hp; self.color = color
    blocks = []
    def generate_level():
        blocks.clear()
        rows = 5 + level // 2
        cols = 8 + level
        for r in range(rows):
            for c in range(cols):
                if c < W-2:
                    hp = random.choices([1,2,3], weights=[70,20,10])[0]
                    color = min(hp+1, 7)
                    blocks.append(Block(c*2+2, r*2+2, hp, color))
    generate_level()

    # Бонусы
    bonuses = []  # (x, y, type) type: 0-wide, 1-speed, 2-life, 3-fire

    def draw():
        stdscr.clear()
        # Границы
        for y in range(H+1):
            stdscr.addch(pad_y+y, pad_x-1, '|')
            stdscr.addch(pad_y+y, pad_x+W, '|')
        for x in range(W+2):
            stdscr.addch(pad_y-1, pad_x+x-1, '-')
        # Платформа
        for i in range(paddle_w):
            stdscr.addch(pad_y+paddle_y, pad_x+paddle_x+i, '=', curses.color_pair(1))
        # Мяч
        if ball_active or not game_over:
            stdscr.addch(pad_y+ball_y, pad_x+ball_x, 'O', curses.color_pair(2))
        # Блоки
        for b in blocks:
            color = curses.color_pair(b.color)
            for i in range(2):
                for j in range(2):
                    stdscr.addch(pad_y+b.y+j, pad_x+b.x+i, '#', color)
        # Бонусы
        for bx, by, btype in bonuses:
            sym = ['W','S','L','F'][btype]
            stdscr.addch(pad_y+by, pad_x+bx, sym, curses.color_pair(3))
        # Счёт
        stdscr.addstr(0, 2, f"Score: {score}", curses.color_pair(4))
        stdscr.addstr(0, W//2-4, f"Best: {best}", curses.color_pair(4))
        stdscr.addstr(0, W-20, f"Lives: {lives}  Level: {level}", curses.color_pair(4))
        if game_over:
            msg = "GAME OVER! Press R to restart, Q to quit"
            stdscr.addstr(H//2, (W - len(msg))//2, msg, curses.color_pair(5))
        stdscr.refresh()

    while True:
        key = stdscr.getch()
        if key == ord('q') or key == ord('Q'): break
        if key == ord('r') or key == ord('R'):
            if game_over:
                paddle_x = (W - paddle_w)//2
                ball_x, ball_y = paddle_x + paddle_w//2, paddle_y - 1
                ball_dx, ball_dy = 1, -1
                ball_active = False
                lives = 3; score = 0; level = 1
                game_over = False
                bonuses.clear()
                generate_level()
                continue
        if key == ord(' '):
            if not ball_active and not game_over:
                ball_active = True
            elif game_over:
                continue
            else:
                # пауза
                pass

        if game_over:
            draw()
            continue

        # Управление платформой
        if key == curses.KEY_LEFT or key == ord('a'):
            paddle_x = max(0, paddle_x - 2)
        elif key == curses.KEY_RIGHT or key == ord('d'):
            paddle_x = min(W - paddle_w, paddle_x + 2)

        # Движение мяча
        if ball_active:
            ball_x += ball_dx * ball_speed
            ball_y += ball_dy * ball_speed
            # Столкновение со стенами
            if ball_x <= 0 or ball_x >= W-1:
                ball_dx *= -1
            if ball_y <= 0:
                ball_dy *= -1
            # Столкновение с платформой
            if ball_y == paddle_y - 1 and paddle_x <= ball_x < paddle_x + paddle_w:
                ball_dy *= -1
                # Угол отскока зависит от положения на платформе
                offset = (ball_x - paddle_x) / paddle_w
                ball_dx = (offset - 0.5) * 2
                if abs(ball_dx) < 0.3:
                    ball_dx = 0.5 if ball_dx >= 0 else -0.5
                # звук
                stdscr.addstr(0,0,'\a')
            # Столкновение с блоками
            for b in blocks[:]:
                if b.x <= ball_x < b.x+2 and b.y <= ball_y < b.y+2:
                    b.hp -= 1
                    if b.hp <= 0:
                        blocks.remove(b)
                        score += 10
                        # бонус
                        if random.random() < 0.15:
                            btype = random.choice([0,1,2,3])
                            bonuses.append((b.x, b.y, btype))
                        stdscr.addstr(0,0,'\a')
                    ball_dy *= -1
                    break
            # Падение мяча
            if ball_y >= H:
                lives -= 1
                if lives <= 0:
                    game_over = True
                    if score > best:
                        best = score
                        save_record(best)
                else:
                    ball_x, ball_y = paddle_x + paddle_w//2, paddle_y - 1
                    ball_dx, ball_dy = 1, -1
                    ball_active = False
            # Уровень пройден
            if not blocks:
                level += 1
                ball_speed += 0.3
                generate_level()
                ball_x, ball_y = paddle_x + paddle_w//2, paddle_y - 1
                ball_dx, ball_dy = 1, -1
                ball_active = False

        # Движение бонусов
        for bonus in bonuses[:]:
            bx, by, btype = bonus
            by += 1
            if by >= H:
                bonuses.remove(bonus)
                continue
            # Проверка касания платформы
            if by == paddle_y and paddle_x <= bx < paddle_x + paddle_w:
                if btype == 0:  # широкая платформа
                    paddle_w = min(16, paddle_w + 4)
                elif btype == 1:  # ускорение
                    ball_speed += 0.5
                elif btype == 2:  # жизнь
                    lives += 1
                elif btype == 3:  # огненный мяч
                    # прожигает блоки насквозь (упрощённо: удваиваем урон)
                    for b in blocks:
                        b.hp -= 1
                bonuses.remove(bonus)
                stdscr.addstr(0,0,'\a')
            else:
                bonuses[bonuses.index(bonus)] = (bx, by, btype)

        draw()
        time.sleep(frame_time)

if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('-s', '--speed', type=int, default=50, help='Скорость (мс)')
    args = parser.parse_args()
    try:
        curses.wrapper(main, args.speed)
    except KeyboardInterrupt:
        print("\nИгра завершена.")
