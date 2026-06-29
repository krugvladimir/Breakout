#!/usr/bin/env ruby
# breakout.rb
# encoding: UTF-8

require 'curses'
require 'json'
require 'fileutils'

RECORD_FILE = File.join(Dir.home, '.breakout_record.json')
def load_record
  return 0 unless File.exist?(RECORD_FILE)
  JSON.parse(File.read(RECORD_FILE))['record'] || 0
rescue
  0
end
def save_record(record)
  File.write(RECORD_FILE, JSON.pretty_generate(record: record))
end

Curses.init_screen
Curses.start_color
Curses.use_default_colors
(1..7).each { |i| Curses.init_pair(i, i, -1) }

height = Curses.lines
width = Curses.cols
if height < 25 || width < 50
  puts "Terminal too small"
  exit 1
end

speed = 50
if ARGV.include?('-s') && ARGV.index('-s') + 1 < ARGV.size
  speed = ARGV[ARGV.index('-s') + 1].to_i
end

H = height - 2
W = width - 2
padY = 1; padX = 1

paddleW = 8
paddleX = (W - paddleW) / 2
paddleY = H - 2
ballX = paddleX + paddleW/2.0
ballY = paddleY - 1
ballDX = 1.0
ballDY = -1.0
ballSpeed = 1.5
ballActive = false
gameOver = false
lives = 3
score = 0
level = 1
best = load_record
frame = speed / 1000.0

blocks = []
bonuses = []

def generate_level(blocks, level, W)
  blocks.clear
  rows = 5 + level/2
  cols = 8 + level
  rows.times do |r|
    cols.times do |c|
      if c*2+2 < W-2
        hp = rand(100) < 70 ? 1 : (rand(100) < 50 ? 2 : 3)
        color = [hp+1, 7].min
        blocks << {x: c*2+2, y: r*2+2, hp: hp, color: color}
      end
    end
  end
end
generate_level(blocks, level, W)

Curses.curs_set(0)
Curses.noecho
Curses.timeout=0

loop do
  ch = Curses.getch
  if ch == 'q' || ch == 'Q'
    break
  elsif ch == 'r' || ch == 'R'
    if gameOver
      paddleX = (W-paddleW)/2
      ballX = paddleX + paddleW/2.0
      ballY = paddleY - 1
      ballDX = 1.0
      ballDY = -1.0
      ballActive = false
      lives = 3
      score = 0
      level = 1
      gameOver = false
      bonuses.clear
      paddleW = 8
      ballSpeed = 1.5
      generate_level(blocks, level, W)
      next
    end
  elsif ch == ' '
    if !ballActive && !gameOver
      ballActive = true
    end
  elsif ch == Curses::KEY_LEFT || ch == 'a'
    paddleX = [0, paddleX - 2].max
  elsif ch == Curses::KEY_RIGHT || ch == 'd'
    paddleX = [W - paddleW, paddleX + 2].min
  end

  if gameOver
    Curses.clear
    msg = "GAME OVER! Score: #{score}  Best: #{best}"
    Curses.setpos(H/2-2, (W - msg.length)/2)
    Curses.attron(Curses.color_pair(4)) { Curses.addstr(msg) }
    Curses.setpos(H/2, (W-20)/2)
    Curses.attron(Curses.color_pair(4)) { Curses.addstr("R - restart | Q - quit") }
    Curses.refresh
    next
  end

  if ballActive
    ballX += ballDX * ballSpeed
    ballY += ballDY * ballSpeed
    if ballX <= 0 || ballX >= W-1
      ballDX *= -1
    end
    if ballY <= 0
      ballDY *= -1
    end
    if ballY.to_i == paddleY-1 && paddleX <= ballX.to_i && ballX.to_i < paddleX+paddleW
      ballDY *= -1
      offset = (ballX - paddleX) / paddleW
      ballDX = (offset - 0.5) * 2
      if ballDX.abs < 0.3
        ballDX = ballDX >= 0 ? 0.5 : -0.5
      end
      print "\a"
    end
    blocks.each do |b|
      if b[:x] <= ballX && ballX < b[:x]+2 && b[:y] <= ballY && ballY < b[:y]+2
        b[:hp] -= 1
        if b[:hp] <= 0
          score += 10
          if rand(100) < 15
            type = rand(4)
            bonuses << {x: b[:x], y: b[:y], type: type}
          end
          blocks.delete(b)
          print "\a"
        end
        ballDY *= -1
        break
      end
    end
    if ballY >= H
      lives -= 1
      if lives <= 0
        gameOver = true
        if score > best
          best = score
          save_record(best)
        end
      else
        ballX = paddleX + paddleW/2.0
        ballY = paddleY - 1
        ballDX = 1.0
        ballDY = -1.0
        ballActive = false
      end
    end
    if blocks.empty?
      level += 1
      ballSpeed += 0.3
      generate_level(blocks, level, W)
      ballX = paddleX + paddleW/2.0
      ballY = paddleY - 1
      ballDX = 1.0
      ballDY = -1.0
      ballActive = false
    end
  end

  bonuses.each do |b|
    b[:y] += 1
    if b[:y] >= H
      bonuses.delete(b)
      next
    end
    if b[:y] == paddleY && paddleX <= b[:x] && b[:x] < paddleX+paddleW
      case b[:type]
      when 0 then paddleW = [16, paddleW+4].min
      when 1 then ballSpeed += 0.5
      when 2 then lives += 1
      when 3
        blocks.each { |bb| bb[:hp] -= 1 }
      end
      bonuses.delete(b)
      print "\a"
    end
  end

  # Draw
  Curses.clear
  (0..H).each do |y|
    Curses.setpos(padY+y, padX-1); Curses.addstr('|')
    Curses.setpos(padY+y, padX+W); Curses.addstr('|')
  end
  (0..W+1).each do |x|
    Curses.setpos(padY-1, padX+x-1); Curses.addstr('-')
  end
  (0...paddleW).each do |i|
    Curses.setpos(padY+paddleY, padX+paddleX+i)
    Curses.attron(Curses.color_pair(1)) { Curses.addstr('=') }
  end
  if ballActive || !gameOver
    Curses.setpos(padY+ballY.to_i, padX+ballX.to_i)
    Curses.attron(Curses.color_pair(2)) { Curses.addstr('O') }
  end
  blocks.each do |b|
    Curses.attron(Curses.color_pair(b[:color]))
    (0...2).each do |i|
      (0...2).each do |j|
        Curses.setpos(padY+b[:y]+j, padX+b[:x]+i)
        Curses.addstr('#')
      end
    end
    Curses.attroff(Curses.color_pair(b[:color]))
  end
  bonuses.each do |b|
    sym = ['W','S','L','F'][b[:type]]
    Curses.setpos(padY+b[:y], padX+b[:x])
    Curses.attron(Curses.color_pair(3)) { Curses.addstr(sym) }
  end
  Curses.attron(Curses.color_pair(4))
  Curses.setpos(0, 2); Curses.addstr("Score: #{score}")
  Curses.setpos(0, W/2-4); Curses.addstr("Best: #{best}")
  Curses.setpos(0, W-20); Curses.addstr("Lives: #{lives}  Level: #{level}")
  Curses.attroff(Curses.color_pair(4))
  if gameOver
    msg = "GAME OVER! Press R to restart, Q to quit"
    Curses.setpos(H/2, (W-msg.length)/2)
    Curses.attron(Curses.color_pair(5)) { Curses.addstr(msg) }
  end
  Curses.refresh
  sleep(frame)
end

Curses.close_screen
