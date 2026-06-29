// breakout.cpp
#include <curses.h>
#include <stdlib.h>
#include <time.h>
#include <unistd.h>
#include <fstream>
#include <string>
#include <vector>
#include <json/json.h>

using namespace std;

int loadRecord() {
    ifstream f(getenv("HOME") + string("/.breakout_record.json"));
    Json::Value root;
    if (f >> root) return root["record"].asInt();
    return 0;
}

void saveRecord(int record) {
    Json::Value root;
    root["record"] = record;
    ofstream f(getenv("HOME") + string("/.breakout_record.json"));
    f << root.toStyledString();
}

struct Block { int x, y, hp, color; };
struct Bonus { int x, y, type; };

int main(int argc, char* argv[]) {
    int speed = 50;
    for (int i=1; i<argc; ++i) {
        if (string(argv[i]) == "-s" && i+1 < argc) speed = atoi(argv[++i]);
        else if (string(argv[i]) == "-h") { cout << "Usage: breakout [-s speed_ms]\n"; return 0; }
    }

    initscr();
    cbreak();
    noecho();
    curs_set(0);
    nodelay(stdscr, TRUE);
    keypad(stdscr, TRUE);
    start_color();
    for (int i=1; i<=7; ++i) init_pair(i, i, COLOR_BLACK);

    int height, width;
    getmaxyx(stdscr, height, width);
    if (height < 25 || width < 50) { endwin(); cout << "Terminal too small\n"; return 1; }

    int H = height-2, W = width-2;
    int pad_y=1, pad_x=1;

    int paddle_w=8, paddle_x=(W-paddle_w)/2, paddle_y=H-2;
    float ball_x=paddle_x+paddle_w/2, ball_y=paddle_y-1;
    float ball_dx=1, ball_dy=-1, ball_speed=1.5;
    bool ball_active=false, game_over=false;
    int lives=3, score=0, level=1;
    int best=loadRecord();
    int frame = speed*1000;

    vector<Block> blocks;
    vector<Bonus> bonuses;

    auto generate_level = [&]() {
        blocks.clear();
        int rows = 5 + level/2;
        int cols = 8 + level;
        for (int r=0; r<rows; ++r) {
            for (int c=0; c<cols; ++c) {
                if (c*2+2 < W-2) {
                    int hp = (rand()%100 < 70) ? 1 : (rand()%100 < 50 ? 2 : 3);
                    int color = min(hp+1, 7);
                    blocks.push_back({c*2+2, r*2+2, hp, color});
                }
            }
        }
    };
    generate_level();

    auto draw = [&]() {
        clear();
        for (int y=0; y<=H; ++y) { mvaddch(pad_y+y, pad_x-1, '|'); mvaddch(pad_y+y, pad_x+W, '|'); }
        for (int x=0; x<W+2; ++x) mvaddch(pad_y-1, pad_x+x-1, '-');
        attron(COLOR_PAIR(1));
        for (int i=0; i<paddle_w; ++i) mvaddch(pad_y+paddle_y, pad_x+paddle_x+i, '=');
        attroff(COLOR_PAIR(1));
        if (ball_active || !game_over) {
            attron(COLOR_PAIR(2));
            mvaddch(pad_y+(int)ball_y, pad_x+(int)ball_x, 'O');
            attroff(COLOR_PAIR(2));
        }
        for (auto &b : blocks) {
            attron(COLOR_PAIR(b.color));
            for (int i=0; i<2; ++i) for (int j=0; j<2; ++j)
                mvaddch(pad_y+b.y+j, pad_x+b.x+i, '#');
            attroff(COLOR_PAIR(b.color));
        }
        attron(COLOR_PAIR(3));
        for (auto &bon : bonuses) {
            char sym = (bon.type==0)?'W':(bon.type==1)?'S':(bon.type==2)?'L':'F';
            mvaddch(pad_y+bon.y, pad_x+bon.x, sym);
        }
        attroff(COLOR_PAIR(3));
        attron(COLOR_PAIR(4));
        mvprintw(0, 2, "Score: %d", score);
        mvprintw(0, W/2-4, "Best: %d", best);
        mvprintw(0, W-20, "Lives: %d  Level: %d", lives, level);
        attroff(COLOR_PAIR(4));
        if (game_over) {
            const char* msg = "GAME OVER! Press R to restart, Q to quit";
            mvprintw(H/2, (W - strlen(msg))/2, "%s", msg);
        }
        refresh();
    };

    while (true) {
        int ch = getch();
        if (ch=='q'||ch=='Q') break;
        if (ch=='r'||ch=='R') {
            if (game_over) {
                paddle_x=(W-paddle_w)/2; ball_x=paddle_x+paddle_w/2; ball_y=paddle_y-1;
                ball_dx=1; ball_dy=-1; ball_active=false;
                lives=3; score=0; level=1; game_over=false;
                bonuses.clear(); paddle_w=8; ball_speed=1.5;
                generate_level();
                continue;
            }
        }
        if (ch==' ') {
            if (!ball_active && !game_over) ball_active = true;
        }

        if (game_over) { draw(); continue; }

        if (ch==KEY_LEFT || ch=='a') paddle_x = max(0, paddle_x-2);
        else if (ch==KEY_RIGHT || ch=='d') paddle_x = min(W-paddle_w, paddle_x+2);

        if (ball_active) {
            ball_x += ball_dx * ball_speed;
            ball_y += ball_dy * ball_speed;
            if (ball_x <= 0 || ball_x >= W-1) ball_dx *= -1;
            if (ball_y <= 0) ball_dy *= -1;
            if ((int)ball_y == paddle_y - 1 && paddle_x <= (int)ball_x && (int)ball_x < paddle_x + paddle_w) {
                ball_dy *= -1;
                float offset = (ball_x - paddle_x) / paddle_w;
                ball_dx = (offset - 0.5) * 2;
                if (abs(ball_dx) < 0.3) ball_dx = (ball_dx >= 0) ? 0.5 : -0.5;
                putchar('\a');
            }
            for (auto it=blocks.begin(); it!=blocks.end(); ++it) {
                if (it->x <= ball_x && ball_x < it->x+2 &&
                    it->y <= ball_y && ball_y < it->y+2) {
                    it->hp--;
                    if (it->hp <= 0) {
                        score += 10;
                        if (rand()%100 < 15) {
                            int btype = rand()%4;
                            bonuses.push_back({it->x, it->y, btype});
                        }
                        blocks.erase(it);
                        putchar('\a');
                    }
                    ball_dy *= -1;
                    break;
                }
            }
            if (ball_y >= H) {
                lives--;
                if (lives <= 0) { game_over=true; if(score>best){best=score; saveRecord(best);} }
                else { ball_x=paddle_x+paddle_w/2; ball_y=paddle_y-1; ball_dx=1; ball_dy=-1; ball_active=false; }
            }
            if (blocks.empty()) {
                level++; ball_speed += 0.3;
                generate_level();
                ball_x=paddle_x+paddle_w/2; ball_y=paddle_y-1;
                ball_dx=1; ball_dy=-1; ball_active=false;
            }
        }

        for (auto it=bonuses.begin(); it!=bonuses.end(); ) {
            it->y += 1;
            if (it->y >= H) { it = bonuses.erase(it); continue; }
            if (it->y == paddle_y && paddle_x <= it->x && it->x < paddle_x + paddle_w) {
                if (it->type == 0) paddle_w = min(16, paddle_w+4);
                else if (it->type == 1) ball_speed += 0.5;
                else if (it->type == 2) lives++;
                else if (it->type == 3) {
                    for (auto &b : blocks) b.hp -= 1;
                }
                it = bonuses.erase(it);
                putchar('\a');
            } else ++it;
        }

        draw();
        usleep(frame);
    }
    endwin();
    return 0;
}
