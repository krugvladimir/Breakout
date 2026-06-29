// breakout.java
import com.googlecode.lanterna.TerminalPosition;
import com.googlecode.lanterna.TerminalSize;
import com.googlecode.lanterna.TextColor;
import com.googlecode.lanterna.graphics.TextGraphics;
import com.googlecode.lanterna.input.KeyStroke;
import com.googlecode.lanterna.terminal.DefaultTerminalFactory;
import com.googlecode.lanterna.terminal.Terminal;
import java.io.*;
import java.nio.file.*;
import java.util.*;
import com.google.gson.*;

public class breakout {
    private static String configFile = System.getProperty("user.home") + "/.breakout_record.json";
    private static int loadRecord() throws IOException {
        Path path = Paths.get(configFile);
        if (!Files.exists(path)) return 0;
        JsonObject obj = new Gson().fromJson(new String(Files.readAllBytes(path)), JsonObject.class);
        return obj.get("record").getAsInt();
    }
    private static void saveRecord(int record) throws IOException {
        JsonObject obj = new JsonObject();
        obj.addProperty("record", record);
        Files.write(Paths.get(configFile), new GsonBuilder().setPrettyPrinting().create().toJson(obj).getBytes());
    }

    static class Block { int x, y, hp, color; }
    static class Bonus { int x, y, type; }

    public static void main(String[] args) throws Exception {
        int speed = 50;
        for (int i=0; i<args.length; i++) {
            if (args[i].equals("-s") && i+1 < args.length) speed = Integer.parseInt(args[++i]);
            else if (args[i].equals("-h")) { System.out.println("Usage: breakout [-s speed_ms]"); return; }
        }
        Terminal terminal = new DefaultTerminalFactory().createTerminal();
        terminal.enterPrivateMode();
        terminal.setCursorVisible(false);
        TerminalSize size = terminal.getTerminalSize();
        int height = size.getRows(), width = size.getColumns();
        if (height < 25 || width < 50) { System.out.println("Terminal too small"); System.exit(1); }

        int H = height-2, W = width-2;
        int padY=1, padX=1;
        int paddleW=8, paddleX=(W-paddleW)/2, paddleY=H-2;
        float ballX=paddleX+paddleW/2f, ballY=paddleY-1;
        float ballDX=1, ballDY=-1, ballSpeed=1.5f;
        boolean ballActive=false, gameOver=false;
        int lives=3, score=0, level=1;
        int best=loadRecord();
        int frame=speed;

        List<Block> blocks = new ArrayList<>();
        List<Bonus> bonuses = new ArrayList<>();
        Random rand = new Random();

        Runnable generateLevel = () -> {
            blocks.clear();
            int rows = 5 + level/2;
            int cols = 8 + level;
            for (int r=0; r<rows; r++) {
                for (int c=0; c<cols; c++) {
                    if (c*2+2 < W-2) {
                        int hp = 1;
                        int rnd = rand.nextInt(100);
                        if (rnd < 70) hp=1;
                        else if (rnd < 90) hp=2;
                        else hp=3;
                        int color = Math.min(hp+1, 7);
                        blocks.add(new Block(){ { x=c*2+2; y=r*2+2; this.hp=hp; this.color=color; } });
                    }
                }
            }
        };
        generateLevel.run();

        TextGraphics tg = terminal.newTextGraphics();

        while (true) {
            KeyStroke key = terminal.pollInput();
            if (key != null) {
                char ch = key.getCharacter() != null ? key.getCharacter() : 0;
                if (ch == 'q' || ch == 'Q') break;
                if (ch == 'r' || ch == 'R') {
                    if (gameOver) {
                        paddleX=(W-paddleW)/2; ballX=paddleX+paddleW/2f; ballY=paddleY-1;
                        ballDX=1; ballDY=-1; ballActive=false;
                        lives=3; score=0; level=1; gameOver=false;
                        bonuses.clear(); paddleW=8; ballSpeed=1.5f;
                        generateLevel.run();
                        continue;
                    }
                }
                if (ch == ' ') {
                    if (!ballActive && !gameOver) ballActive = true;
                }
                if (key.getKeyType() == KeyStroke.KeyType.ArrowLeft || ch == 'a') paddleX = Math.max(0, paddleX-2);
                if (key.getKeyType() == KeyStroke.KeyType.ArrowRight || ch == 'd') paddleX = Math.min(W-paddleW, paddleX+2);
            }

            if (gameOver) {
                tg.clear();
                String msg = "GAME OVER! Score: " + score + "  Best: " + best;
                tg.putString((width - msg.length())/2, height/2-2, msg, TextColor.ANSI.RED);
                tg.putString((width - 20)/2, height/2, "R - restart | Q - quit", TextColor.ANSI.CYAN);
                terminal.flush();
                continue;
            }

            if (ballActive) {
                ballX += ballDX * ballSpeed;
                ballY += ballDY * ballSpeed;
                if (ballX <= 0 || ballX >= W-1) ballDX *= -1;
                if (ballY <= 0) ballDY *= -1;
                if ((int)ballY == paddleY-1 && paddleX <= (int)ballX && (int)ballX < paddleX+paddleW) {
                    ballDY *= -1;
                    float offset = (ballX - paddleX) / paddleW;
                    ballDX = (offset - 0.5f) * 2;
                    if (Math.abs(ballDX) < 0.3f) ballDX = ballDX >= 0 ? 0.5f : -0.5f;
                    System.out.print("\007");
                }
                for (Iterator<Block> it = blocks.iterator(); it.hasNext(); ) {
                    Block b = it.next();
                    if (b.x <= ballX && ballX < b.x+2 && b.y <= ballY && ballY < b.y+2) {
                        b.hp--;
                        if (b.hp <= 0) {
                            score += 10;
                            if (rand.nextInt(100) < 15) {
                                int type = rand.nextInt(4);
                                bonuses.add(new Bonus(){ { x=b.x; y=b.y; this.type=type; } });
                            }
                            it.remove();
                            System.out.print("\007");
                        }
                        ballDY *= -1;
                        break;
                    }
                }
                if (ballY >= H) {
                    lives--;
                    if (lives <= 0) { gameOver=true; if(score>best){best=score; saveRecord(best);} }
                    else { ballX=paddleX+paddleW/2f; ballY=paddleY-1; ballDX=1; ballDY=-1; ballActive=false; }
                }
                if (blocks.isEmpty()) {
                    level++; ballSpeed += 0.3f;
                    generateLevel.run();
                    ballX=paddleX+paddleW/2f; ballY=paddleY-1;
                    ballDX=1; ballDY=-1; ballActive=false;
                }
            }

            for (Iterator<Bonus> it = bonuses.iterator(); it.hasNext(); ) {
                Bonus b = it.next();
                b.y++;
                if (b.y >= H) { it.remove(); continue; }
                if (b.y == paddleY && paddleX <= b.x && b.x < paddleX+paddleW) {
                    if (b.type == 0) paddleW = Math.min(16, paddleW+4);
                    else if (b.type == 1) ballSpeed += 0.5f;
                    else if (b.type == 2) lives++;
                    else if (b.type == 3) { for (Block bb : blocks) bb.hp -= 1; }
                    it.remove();
                    System.out.print("\007");
                }
            }

            tg.clear();
            for (int y=0; y<=H; y++) {
                tg.putString(padX-1, padY+y, "|", TextColor.ANSI.WHITE);
                tg.putString(padX+W, padY+y, "|", TextColor.ANSI.WHITE);
            }
            for (int x=0; x<W+2; x++) tg.putString(padX+x-1, padY-1, "-", TextColor.ANSI.WHITE);
            for (int i=0; i<paddleW; i++) tg.putString(padX+paddleX+i, padY+paddleY, "=", TextColor.ANSI.YELLOW);
            if (ballActive || !gameOver) tg.putString(padX+(int)ballX, padY+(int)ballY, "O", TextColor.ANSI.CYAN);
            for (Block b : blocks) {
                for (int i=0; i<2; i++) for (int j=0; j<2; j++)
                    tg.putString(padX+b.x+i, padY+b.y+j, "#", TextColor.ANSI.valueOf(b.color));
            }
            for (Bonus b : bonuses) {
                char sym = b.type==0?'W':b.type==1?'S':b.type==2?'L':'F';
                tg.putString(padX+b.x, padY+b.y, String.valueOf(sym), TextColor.ANSI.MAGENTA);
            }
            tg.putString(2, 0, "Score: " + score, TextColor.ANSI.WHITE);
            tg.putString(W/2-4, 0, "Best: " + best, TextColor.ANSI.WHITE);
            tg.putString(W-20, 0, "Lives: " + lives + "  Level: " + level, TextColor.ANSI.WHITE);
            if (gameOver) {
                String msg = "GAME OVER! Press R to restart, Q to quit";
                tg.putString((W - msg.length())/2, H/2, msg, TextColor.ANSI.RED);
            }
            terminal.flush();
            Thread.sleep(frame);
        }
        terminal.exitPrivateMode();
        terminal.close();
    }
}
