# 双人贪吃蛇 - 架构说明

这是一个 Go 写的终端双人对战贪吃蛇游戏，用 termbox-go 做终端 UI。红蓝两条蛇同屏抢食，撞到自己/对方/墙/尸体就死。主打一个怀旧手感，在命令行里就能玩。

---

## 目录结构

```
project178/
├── main.go          # 程序入口，只负责初始化和事件循环
├── game.go          # 核心游戏逻辑（tick、移动、碰撞、拾取）
├── entities.go      # 所有数据结构定义 + 简单方法
├── render.go        # 终端渲染（termbox 绘制全在这）
├── input.go         # 键盘输入处理
├── maps.go          # 3张地图配置 + 蛇的初始构造
├── go.mod           # Go 模块依赖
└── ARCHITECTURE.md  # 就是你现在看的这个
```

---

## 每个文件干啥的

### main.go - 入口 & 调度

**职责**：啥业务逻辑都没有，就管初始化 termbox、开两个 goroutine（键盘监听 + ticker）、然后根据状态机调不同的 handler。

**关键逻辑**：
- 事件队列 `eventQueue` 跑在独立 goroutine 里，把 termbox 的键盘事件塞进来，不阻塞渲染
- 30ms ticker 负责刷新：菜单状态下只渲染菜单；对战状态下先判断距离上一个 tick 够不够 120ms（正常速度），够了就 `Tick()` 推进一步
- 所有状态通过一个 `*Game` 对象流转，没有全局变量

**核心片段** [main.go L12-L65](file:///d:/code/ai-prompt/solo-chrome-dev-F12/repos/repo178/project178/main.go#L12-L65)：
```go
game := NewMenu()
eventQueue := make(chan termbox.Event, 64)
go func() {
    for { eventQueue <- termbox.PollEvent() }
}()
ticker := time.NewTicker(30 * time.Millisecond)
// select 里面分发事件
```

---

### entities.go - 数据模型层

**职责**：所有类型定义、常量、枚举，以及 Snake/Game 的纯方法（不修改外部状态、不涉及 I/O）。

**关键结构体**：

| 类型 | 作用 |
|---|---|
| `Point` | 格子坐标 {X, Y} |
| `Snake` | 一条蛇的全部状态：身体坐标数组、方向、分数、是否存活、各种计时器结束时间 |
| `Food` | 普通食物（+1/+3 分） |
| `Star` | 星星道具（2秒穿墙） |
| `Diamond` | 钻石道具（+5分 + 长3节，场上同时只一个） |
| `Corpse` | 死亡后的尸体块，3秒内是障碍 |
| `FlashMsg` | 顶部闪字提示（吃钻石时用） |
| `Game` | 整个游戏状态机，持有上面所有东西 + 当前状态 + 最后tick时间 |

**状态枚举** `GameState` [entities.go L79-L86](file:///d:/code/ai-prompt/solo-chrome-dev-F12/repos/repo178/project178/entities.go#L79-L86)：
```
StateMenu → StatePlaying ↔ StatePaused
                ↓
            StateGameOver
```

**纯方法**（不会改数据，只是查状态）：
- `Snake.Head()` - 取蛇头坐标
- `Snake.HasStar()` - 穿墙效果是否在有效期
- `Snake.HasBoost()` - 加速效果是否在有效期
- `Game.IsWall()` - 判断某格是不是墙（边界或障碍物）

---

### game.go - 核心游戏逻辑

**职责**：游戏怎么玩全在这。状态推进、碰撞检测、拾取判定、生成器。

**构造器**：
- `NewGame(mapID)` - 新建一局对战，初始化两条蛇、两个食物、概率生成钻石
- `NewMenu()` - 新建菜单状态

**每帧逻辑** `Tick()` [game.go L262-L286](file:///d:/code/ai-prompt/solo-chrome-dev-F12/repos/repo178/project178/game.go#L262-L286)：
```
1. 移动 P1 蛇 → 检测碰撞 → 检查拾取
2. 移动 P2 蛇 → 检测碰撞 → 检查拾取
3. 清理过期尸体
4. 清理过期闪字
5. 检查是否双方都死了 → 进入 GameOver
```

**蛇移动** `MoveSnake()` [game.go L164-L216](file:///d:/code/ai-prompt/solo-chrome-dev-F12/repos/repo178/project178/game.go#L164-L216)：
1. 根据方向算出下一格位置
2. 如果有穿墙 buff，碰到边界就从对面出来（经典 wrap-around），碰到障碍物就跳过
3. 碰撞检测：撞墙/撞自己/撞对方/撞尸体 → 死亡，生成尸体块
4. 没死就把新头插到数组前面，尾巴砍掉（或者留着如果要 grow）

**碰撞检测** `CheckCollision()` [game.go L135-L162](file:///d:/code/ai-prompt/solo-chrome-dev-F12/repos/repo178/project178/game.go#L135-L162)：
- 有穿墙 buff 时跳过墙的判定
- 注意：蛇的最后一格（尾巴）如果不 grow 的话下一帧会消失，所以不算碰撞

**拾取判定** `CheckPickups()` [game.go L218-L260](file:///d:/code/ai-prompt/solo-chrome-dev-F12/repos/repo178/project178/game.go#L218-L260)：
- 吃普通食物 → 加分 + 加长 + 概率刷新额外食物/星星/钻石
- 吃星星 → StarEnd = now + 2s
- 吃钻石 → +5分 + 长3节 + 闪字提示 + 立刻在新位置重生钻石

**生成器**：
- `SpawnFood()` - 随机空位放食物，1/5概率是+3分的
- `SpawnStar()` - 随机空位放星星
- `SpawnDiamond()` - 随机空位放钻石（场上只有一个，由 `Diamond != nil` 保证）
- `Occupied()` - 判断某格有没有被任何东西占（墙/蛇/食物/道具/尸体）
- `RandomEmpty()` - 找一个空位，试500次找不到就放弃

**暂停时间补偿** `shiftTimers()` [game.go L300-L326](file:///d:/code/ai-prompt/solo-chrome-dev-F12/repos/repo178/project178/game.go#L300-L326)：
暂停恢复时把所有时间戳（StarEnd/BoostEnd/CorpseEnd/Flash.ExpiresAt/LastTick）统一往后顺延暂停时长。
因为这些计时器都是墙钟时间戳（`time.Now() + duration`），暂停时墙钟还在走，所以恢复时要补回来。

---

### render.go - 终端渲染

**职责**：所有 `termbox.SetCell` 调用都在这，游戏逻辑不掺和。

**帧率**：由 main.go 的 30ms ticker 驱动，≈33 FPS。

**绘制单位** `drawCell()` [render.go L24-L28](file:///d:/code/ai-prompt/solo-chrome-dev-F12/repos/repo178/project178/render.go#L24-L28)：
因为终端字符是长方形，一个格子占 2 个字符宽度才接近正方形。所以每个游戏格子实际画两个终端字符（字符 + 空格）。

**主渲染** `Render()` [render.go L30-L135](file:///d:/code/ai-prompt/solo-chrome-dev-F12/repos/repo178/project178/render.go#L30-L135)：
```
1. 清屏
2. 顶部分数栏（吃钻石时会闪烁变色 + 右侧闪字）
3. 操作提示栏
4. 画边框（┌─┐││└─┘）
5. 画障碍物（█）
6. 画尸体（x，对应蛇的颜色）
7. 画食物（● 1分 / ◆ 3分，绿色）
8. 画星星（★，黄色）
9. 画钻石（♦，亮黄色）
10. 画两条蛇（■身体 ●头，穿墙时头变成◎ + 紫色闪烁背景）
11. 暂停/结束弹窗
12. termbox.Flush() 输出到终端
```

**蛇绘制** `renderSnake()` [render.go L137-L161](file:///d:/code/ai-prompt/solo-chrome-dev-F12/repos/repo178/project178/render.go#L137-L161)：
- 蛇头和身体区分颜色，蛇头更亮
- 穿墙激活时蛇头用 `◎` 符号，偶数节身体背景变紫色闪烁

**菜单渲染** `RenderMenu()` [render.go L194-L262](file:///d:/code/ai-prompt/solo-chrome-dev-F12/repos/repo178/project178/render.go#L194-L262)：
- 标题、副标题、地图选择列表（选中项高亮）、操作说明、规则说明

**弹窗**：
- `renderOverlay()` - 通用弹窗（暂停用）
- `renderGameOver()` - 游戏结束弹窗（显示胜负和分数）

---

### input.go - 键盘输入

**职责**：把 termbox 的键盘事件翻译成游戏操作。两个方法：`HandleKeyMenu` 和 `HandleKeyPlaying`。

**菜单按键** `HandleKeyMenu()` [input.go L9-L29](file:///d:/code/ai-prompt/solo-chrome-dev-F12/repos/repo178/project178/input.go#L9-L29)：
- ↑↓ 选地图
- Enter / Space 开始游戏
- Esc 退出

**对战按键** `HandleKeyPlaying()` [input.go L31-L89](file:///d:/code/ai-prompt/solo-chrome-dev-F12/repos/repo178/project178/input.go#L31-L89)：

| 按键 | 玩家 | 作用 |
|---|---|---|
| W A S D | P1 (红) | 上左下右移动（不能直接反向） |
| Q | P1 (红) | 加速 250ms |
| ↑↓←→ | P2 (蓝) | 移动 |
| Space | P2 (蓝) | 加速 250ms |
| P | 共同 | 暂停/恢复（恢复时调用 shiftTimers 补偿时间） |
| R | 共同 | 重开（`*g = *NewGame(g.MapID)` 整个替换） |
| M | 共同 | 回菜单（`*g = *NewMenu()`） |
| Esc | 共同 | 退出 |

---

### maps.go - 地图配置

**职责**：3张地图的障碍物坐标，以及 `NewSnake()` 蛇的初始构造函数。

**地图数据** `mapConfigs` [maps.go L12-L45](file:///d:/code/ai-prompt/solo-chrome-dev-F12/repos/repo178/project178/maps.go#L12-L45)：
每张地图就是一个 `[]Point`，存障碍物坐标。地图是 24×16 格子。

1. **Cross Roads**（十字路口）：简单的 3 块障碍，适合新手
2. **Cornered**（围城）：角落有长廊障碍，增加路线复杂度
3. **Fortress**（堡垒）：最复杂，四角和中央都有障碍，容易被逼死

**蛇初始构造** `NewSnake()` [maps.go L54-L80](file:///d:/code/ai-prompt/solo-chrome-dev-F12/repos/repo178/project178/maps.go#L54-L80)：
- P1 初始在左边 (5, 8)，朝右，3节长
- P2 初始在右边 (18, 8)，朝左，3节长

---

## 游戏状态机

```
                        Enter/Space
  ┌──────────┐    ┌──────────────────────┐
  │  Menu    │───▶│      Playing         │
  └──────────┘    └──────────────────────┘
                       ▲          │
                       │ P        │ 双方都死
                       │          ▼
                  ┌─────────┐  ┌───────────┐
                  │ Paused  │  │ GameOver  │
                  └─────────┘  └───────────┘
```

**状态切换**：
- `Menu → Playing`：选好地图按 Enter/Space，`*g = *NewGame(selected)`
- `Playing → Paused`：按 P，`g.State = StatePaused`，记录 `PausedAt`
- `Paused → Playing`：按 P，调用 `shiftTimers()` 补偿时间
- `Playing → GameOver`：Tick 中检测到双方都死了
- `Any → Menu`：按 M，`*g = *NewMenu()`
- `Any → Restart`：按 R，`*g = *NewGame(g.MapID)` 整个对象替换

---

## 事件处理

### 死亡判定（`CheckCollision` + `MoveSnake`）
1. 没有穿墙 buff 时，撞墙（边界或障碍物）→ 死
2. 撞自己身体（尾巴不算，因为马上要走掉）→ 死
3. 撞对方身体 → 死
4. 撞尸体（`x` 标记，3秒内）→ 死

死亡后：
- `s.Alive = false`
- 生成对应颜色的尸体块（`Corpse`），3秒后消失
- 如果是单人死亡，另一个可以继续吃分，直到双方都死才 GameOver

### 吃食物（`CheckPickups`）
- +1 分（●）/ +3 分（◆），食物吃完立刻在新位置刷新
- 额外有小概率多刷一个食物 / 刷星星 / 刷钻石（如果还没有）

### 吃星星（`CheckPickups`）
- `StarEnd = time.Now() + 2s`
- 期间穿墙、不检测墙壁碰撞、出边界 wrap-around

### 吃钻石（`CheckPickups`）
- +5 分，长 3 节
- 触发顶部闪字 "P1/P2 got ♦ +5!"，分数栏闪烁
- 钻石立刻在新空位重生（保证场上永远只有一个，而且刺激感够）

### 暂停恢复（`HandleKeyPlaying` 里的 P 键）
- 暂停时记录 `PausedAt = time.Now()`
- 恢复时 `pauseDur = time.Since(PausedAt)`，调用 `shiftTimers(pauseDur)`
- 所有时间戳统一往后顺延，防止暂停期间效果过期

---

## 坑 & TODO（以后回来改）

### 已知写得不够好的地方

1. **`*g = *NewGame(...)` 整体替换对象** [input.go L84/L86](file:///d:/code/ai-prompt/solo-chrome-dev-F12/repos/repo178/project178/input.go#L84-L86)
   - 简单粗暴但有效，不过 `Game` 里有指针字段（`Player1`/`Player2` 都是指针），这种浅拷贝是 OK 的（因为每次 NewGame 都新建了指针），但以后如果在 Game 里加了更多指针要小心
   - 以后可以改成 `g.Reset()` 方法，显式重置每个字段，更清晰也避免拷贝遗漏

2. **tick 累积问题**
   - `now.Sub(LastTick) >= tr` 这种判断，在系统卡顿后可能一帧内触发多次 tick（正常行为，追帧），但如果卡死很久恢复后可能一次性把几百个 tick 跑完，蛇直接飞出屏幕
   - 可以加个最大追帧限制，比如一次最多追 3 个 tick

3. **钻石生成概率**
   - 现在 `DiamondSpawnChance = 8` 是 12.5% 概率，吃食物时触发。感觉还行但如果觉得太多或太少可以调
   - 场上永远只存在一个钻石，这个机制不错，不用改

4. **尸体残留的碰撞问题**
   - 现在 `Occupied()` 把尸体算占用，`CheckCollision` 也检查尸体，没问题
   - 但尸体 3 秒后消失的逻辑是在 `Tick()` 里按墙钟走的，暂停时也会用 `shiftTimers` 顺延，这块应该是对的

5. **`HasStar()` / `HasBoost()` 用墙钟判断** [entities.go L111-L117](file:///d:/code/ai-prompt/solo-chrome-dev-F12/repos/repo178/project178/entities.go#L111-L117)
   - 之前的 bug 就是因为暂停时墙钟还在走，现在加了 `shiftTimers` 补偿应该没问题了
   - 但如果以后想加个游戏内时间（和墙钟完全脱离），可以把所有 `time.Now()` 换成 `g.GameTime`，暂停时就不用补偿了，更干净

### 可以加的功能

- [ ] AI 玩家（单人模式打电脑）
- [ ] 更多地图（现在 3 张，可以加到 5 张）
- [ ] 更多道具：减速对方、缩短对方、无敌（不是穿墙，是真无敌）
- [ ] 音效（termbox 做不了，得换 beep 之类的库）
- [ ] 分数记录（存文件，排行榜）
- [ ] 网络对战（TCP/WebSocket）
- [ ] 用游戏内时间替代墙钟，消除 `shiftTimers` 这种补偿逻辑
- [ ] 重放功能（把每帧输入存下来，结束后可以回放）

---

## 构建 & 运行

```bash
go build -o snakes.exe .
./snakes.exe
```

依赖只有 `github.com/nsf/termbox-go`，纯终端 UI，不需要图形环境。

如果你几个月后回来改代码，从 `main.go` 开始看起，顺着事件循环和状态机就能摸到所有逻辑了。祝玩得开心 🐍
