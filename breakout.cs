// breakout.cs
using System;
using System.Collections.Generic;
using System.IO;
using System.Text.Json;
using System.Threading;
using System.Runtime.InteropServices;

class Breakout
{
    static string Colorize(string text, string color)
    {
        string col = color switch
        {
            "yellow" => "\x1b[93m",
            "cyan" => "\x1b[96m",
            "magenta" => "\x1b[95m",
            "white" => "\x1b[97m",
            "red" => "\x1b[91m",
            _ => "\x1b[0m"
        };
        return col + text + "\x1b[0m";
    }

    static string ConfigFile => Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.UserProfile), ".breakout_record.json");
    static int LoadRecord()
    {
        if (!File.Exists(ConfigFile)) return 0;
        var data = JsonSerializer.Deserialize<Dictionary<string,int>>(File.ReadAllText(ConfigFile));
        return data.GetValueOrDefault("record", 0);
    }
    static void SaveRecord(int record)
    {
        var data = new Dictionary<string,int>{ {"record", record} };
        File.WriteAllText(ConfigFile, JsonSerializer.Serialize(data));
    }

    class Block { public int x, y, hp, color; }
    class Bonus { public int x, y, type; }

    static void Main(string[] args)
    {
        int speed = 50;
        for (int i=0; i<args.Length; i++)
        {
            if (args[i] == "-s" && i+1 < args.Length) speed = int.Parse(args[++i]);
            else if (args[i] == "-h") { Console.WriteLine("Usage: breakout [-s speed_ms]"); return; }
        }
        Console.Clear();
        int height = Console.WindowHeight;
        int width = Console.WindowWidth;
        if (height < 25 || width < 50) { Console.WriteLine("Terminal too small"); return; }
        Random rand = new Random();
        int H = height-2, W = width-2;
        int padY=1, padX=1;

        int paddleW=8;
        int paddleX=(W-paddleW)/2;
        int paddleY=H-2;
        float ballX=paddleX+paddleW/2, ballY=paddleY-1;
        float ballDX=1, ballDY=-1, ballSpeed=1.5f;
        bool ballActive=false, gameOver=false;
        int lives=3, score=0, level=1;
        int best=LoadRecord();
        int frame=speed;

        List<Block> blocks = new List<Block>();
        List<Bonus> bonuses = new List<Bonus>();

        void GenerateLevel()
        {
            blocks.Clear();
            int rows = 5 + level/2;
            int cols = 8 + level;
            for (int r=0; r<rows; r++)
            {
                for (int c=0; c<cols; c++)
                {
                    if (c*2+2 < W-2)
                    {
                        int hp = 1;
                        int rnd = rand.Next(100);
                        if (rnd < 70) hp=1;
                        else if (rnd < 90) hp=2;
                        else hp=3;
                        int color = Math.Min(hp+1, 7);
                        blocks.Add(new Block { x=c*2+2, y=r*2+2, hp=hp, color=color });
                    }
                }
            }
        }
        GenerateLevel();

        Console.CursorVisible = false;
        while (true)
        {
            if (Console.KeyAvailable)
            {
                var key = Console.ReadKey(true).Key;
                if (key == ConsoleKey.Q) break;
                if (key == ConsoleKey.R && gameOver)
                {
                    paddleX=(W-paddleW)/2; ballX=paddleX+paddleW/2; ballY=paddleY-1;
                    ballDX=1; ballDY=-1; ballActive=false;
                    lives=3; score=0; level=1; gameOver=false;
                    bonuses.Clear(); paddleW=8; ballSpeed=1.5f;
                    GenerateLevel();
                    continue;
                }
                if (key == ConsoleKey.Spacebar)
                {
                    if (!ballActive && !gameOver) ballActive = true;
                }
                if (key == ConsoleKey.LeftArrow || key == ConsoleKey.A) paddleX = Math.Max(0, paddleX-2);
                if (key == ConsoleKey.RightArrow || key == ConsoleKey.D) paddleX = Math.Min(W-paddleW, paddleX+2);
            }

            if (gameOver)
            {
                Console.Clear();
                string msg = $"GAME OVER! Score: {score}  Best: {best}";
                Console.SetCursorPosition((width - msg.Length)/2, height/2-2);
                Console.Write(Colorize(msg, "red"));
                Console.SetCursorPosition((width-20)/2, height/2);
                Console.Write(Colorize("R - restart | Q - quit", "cyan"));
                continue;
            }

            if (ballActive)
            {
                ballX += ballDX * ballSpeed;
                ballY += ballDY * ballSpeed;
                if (ballX <= 0 || ballX >= W-1) ballDX *= -1;
                if (ballY <= 0) ballDY *= -1;
                if ((int)ballY == paddleY-1 && paddleX <= (int)ballX && (int)ballX < paddleX+paddleW)
                {
                    ballDY *= -1;
                    float offset = (ballX - paddleX) / paddleW;
                    ballDX = (offset - 0.5f) * 2;
                    if (Math.Abs(ballDX) < 0.3f) ballDX = ballDX >= 0 ? 0.5f : -0.5f;
                    Console.Beep();
                }
                for (int i=0; i<blocks.Count; i++)
                {
                    var b = blocks[i];
                    if (b.x <= ballX && ballX < b.x+2 && b.y <= ballY && ballY < b.y+2)
                    {
                        b.hp--;
                        if (b.hp <= 0)
                        {
                            score += 10;
                            if (rand.Next(100) < 15)
                            {
                                int type = rand.Next(4);
                                bonuses.Add(new Bonus { x=b.x, y=b.y, type=type });
                            }
                            blocks.RemoveAt(i);
                            Console.Beep();
                        }
                        ballDY *= -1;
                        break;
                    }
                }
                if (ballY >= H)
                {
                    lives--;
                    if (lives <= 0) { gameOver=true; if(score>best){best=score; SaveRecord(best);} }
                    else { ballX=paddleX+paddleW/2; ballY=paddleY-1; ballDX=1; ballDY=-1; ballActive=false; }
                }
                if (blocks.Count == 0)
                {
                    level++; ballSpeed += 0.3f;
                    GenerateLevel();
                    ballX=paddleX+paddleW/2; ballY=paddleY-1;
                    ballDX=1; ballDY=-1; ballActive=false;
                }
            }

            for (int i=0; i<bonuses.Count; i++)
            {
                var bon = bonuses[i];
                bon.y++;
                if (bon.y >= H) { bonuses.RemoveAt(i); i--; continue; }
                if (bon.y == paddleY && paddleX <= bon.x && bon.x < paddleX+paddleW)
                {
                    if (bon.type == 0) paddleW = Math.Min(16, paddleW+4);
                    else if (bon.type == 1) ballSpeed += 0.5f;
                    else if (bon.type == 2) lives++;
                    else if (bon.type == 3) { foreach (var b in blocks) b.hp -= 1; }
                    bonuses.RemoveAt(i);
                    i--;
                    Console.Beep();
                }
            }

            // Draw
            Console.Clear();
            for (int y=0; y<=H; y++) { Console.SetCursorPosition(padX-1, padY+y); Console.Write('|'); Console.SetCursorPosition(padX+W, padY+y); Console.Write('|'); }
            for (int x=0; x<W+2; x++) { Console.SetCursorPosition(padX+x-1, padY-1); Console.Write('-'); }
            for (int i=0; i<paddleW; i++) { Console.SetCursorPosition(padX+paddleX+i, padY+paddleY); Console.Write(Colorize("=","yellow")); }
            if (ballActive || !gameOver) { Console.SetCursorPosition(padX+(int)ballX, padY+(int)ballY); Console.Write(Colorize("O","cyan")); }
            foreach (var b in blocks)
            {
                string col = b.color <= 3 ? "cyan" : "white";
                for (int i=0; i<2; i++) for (int j=0; j<2; j++)
                { Console.SetCursorPosition(padX+b.x+i, padY+b.y+j); Console.Write(Colorize("#", col)); }
            }
            foreach (var bon in bonuses)
            {
                char sym = bon.type==0?'W':bon.type==1?'S':bon.type==2?'L':'F';
                Console.SetCursorPosition(padX+bon.x, padY+bon.y);
                Console.Write(Colorize(sym.ToString(), "magenta"));
            }
            Console.SetCursorPosition(2,0); Console.Write(Colorize($"Score: {score}", "white"));
            Console.SetCursorPosition(W/2-4,0); Console.Write(Colorize($"Best: {best}", "white"));
            Console.SetCursorPosition(W-20,0); Console.Write(Colorize($"Lives: {lives}  Level: {level}", "white"));
            if (gameOver)
            {
                string msg = "GAME OVER! Press R to restart, Q to quit";
                Console.SetCursorPosition((W - msg.Length)/2, H/2);
                Console.Write(Colorize(msg, "red"));
            }
            Thread.Sleep(frame);
        }
    }
}
