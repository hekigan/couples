## 🎮 Gameplay Logic

### 1. Users & Rooms

- to start, a user must choose a username
- A friend invitation can be **accepted or declined**
- Once accepted, both users appear in each other’s **friend list**
- A user can create a **room** and invite one friend
- A room can only contain **2 users**

### 2. Starting a Game

- Both users select question categories via checkboxes (the choice is made with checkbox buttons and are updated in real time so that both users can see the choices):
  - Couples / Friends / Sex / Family / etc.
- The room owner clicks **Start Game**
- Randomly decide who starts first
- Supabase Realtime broadcasts room/game state to both users instantly

### 3. Gameplay Flow

- The current player draws a random question (filtered by selected categories)
- Both users see the same question (realtime update)
- The active player can:
  - **Answer** or **Pass**
  - Optionally write their answer (text field)
  - Click **OK** to submit
- The second player:
  - Sees the submitted answer in real-time (after the first player has clicked **OK**)
  - Clicks **Next Card** to draw a new question (and their turn begins)
- Game continues turn by turn
- **Finish Game** button ends the session

### 4. Persistence & History

- Question/Answer history stored by:
  - Room ID
  - Player 1 / Player 2
  - Timestamp
- History ensures that:
  - Previously asked questions are not repeated in future sessions
  - Past sessions can be resumed (if users are registered)
- For anonymous users:
  - Data is session-based
  - History and temporary profile are deleted after the session ends