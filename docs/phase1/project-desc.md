Project Name : Fisher Timer
Group Members :
a. Chanatda Konchom 6631305321
b. Ittichet Thongsang 6631363721
c. Napat Srisamut 6631314021
d. Naphat Serirak 6632061321
e. Chayut Archamongkol 6631307621

Problem Description:
Many learners struggle to maintain focus and motivation when studying alone. Fisher Timer addresses this pain point by offering a community-based study timer system that combines individual time management with real-time virtual study sessions. Users can configure personal work and rest intervals while joining shared rooms, leveraging the positive pressure of studying alongside peers. To sustain long-term engagement, the platform incorporates a fishing-themed gamification system. Users earn rewards for completing focused sessions, track their personal statistics via dashboards, and compare their progress on community leaderboards.
Target Customer:
Customer: Create and join study rooms, configure timers, earn rewards, view personal dashboard and view study session leaderboard.
Admin: Monitor real-time active sessions, review user reports, ban/unban users.

Scenario (use-case & description)
Use Case Name:
Create Study Session Room
ID: UC-01
Important Level: High
Primary Actor: Customer (authenticated user)
Use Case Type: Detail, Essential
Stakeholders and Interests:
Customer : wants a room to study in, with control over how many people can join, including studying alone by setting the limit to 1.
Admin : wants every created room registered so it appears in real-time monitoring.
Brief Description:
An authenticated user creates a study session room. Every room is publicly listed. The creator sets a participant limit between 1 and the system maximum, where 1 means a solo session no one else can join. The creator is automatically enrolled as the first participant. Each participant runs an independent timer inside the room.
Trigger: User selects "Create Room" from the room list.
Type: External
Relationships:
Owned by Study session service.
Requires ban state from AdminModeration
Publishes session.created to Admin monitoring.
Precondition:
User is signed in; gateway has validated a non-expired JWT.
User has no active ban.
User is not currently an active participant in any room.
Postcondition:
StudySession created with status = ACTIVE, creator_id, participant_limit, created_at.
SessionParticipant created for the creator.
Room appears in the public room list and in Admin monitoring.
session.created published.
Normal Flow of Events:
User selects Create Study Session Room.
System displays the create form: room name and participant limit, defaulting to the system maximum.
User enters a room name and participant limit.
User submits.
System verifies the user has no active ban (S-1).
System verifies the user is not already in a room (E-2).
System validates room name and participant limit (E-1).
System creates the StudySession with the creator as owner.
System creates the SessionParticipant record for the creator.
System publishes session.created.
System places the user in the room and loads their saved TimerSetting.
Room becomes visible in the public room list.
Subflows:
S-1 Verify Ban Status: Study session checks the user's ban state. If an active ban exists, the request is rejected and the ban reason and expiry are shown.
Alternate/Exceptional Flow:
E-1 Room name empty or over max length, or participant limit outside MAX_PARTICIPANTS_PER_ROOM. System shows a validation error and returns to step 2.
E-2 User is already an active participant in a room. System offers to leave that room (UC-03) or cancel.
E-3 Persistence failure. No partial record is retained, an error is shown, and the user may retry.
E-4 Missing or expired JWT. Gateway rejects the request and redirects to Sign In.
E-5 User has an active ban. Creation is denied.

Use Case Name:
Join Study Session Room
ID: UC-02
Important Level: High
Primary Actor: Customer (authenticated user)
Use Case Type: Detail, Essential
Stakeholders and Interests:
Customer : wants to study alongside others for accountability, while keeping full control of a personal timer.
Admin : wants accurate live participant lists per room for monitoring.
Brief Description:
An authenticated user browses the public room list and joins an active room that has open capacity. Joining grants presence in the room and visibility of other participants' timer states. It does not link or synchronize timers; each participant continues to run their own.
Trigger: User selects "Join" on a room in the public room list.
Type: External
Relationships:
Owned by Study session service.
Requires ban state from Admin service.
Publishes participant.joined to Admin monitoring.
Precondition:
User is signed in; gateway has validated a non-expired JWT.
User has no active ban.
User is not currently an active participant in any room.
Target StudySession exists with status = ACTIVE.
Current participant count is below the room's participant_limit.
Postcondition:
SessionParticipant created with session_id, user_id, joined_at.
Live participant list updated for all room members and in Admin monitoring.
participant.joined published.
User's timer is available for this room, keyed by (session_id, user_id), in a stopped state.
Normal Flow of Events:
User opens the study session room list.
System displays active rooms with name, current participant count, and capacity.
User selects a room and confirms Join.
System verifies the user has no active ban (S-1).
System verifies the room is still ACTIVE (E-1).
System verifies the room has remaining capacity (E-2).
System verifies the user is not already in a room (E-3).
System creates the SessionParticipant record.
System publishes participant.joined.
System pushes the updated participant list to all room members in real time.
System renders the room with the user's timer stopped, using their saved TimerSetting.
User may start their timer (UC-XXXX). Each completed work cycle awards a reward independently of this use case.
Subflows:
S-1 Verify ban status. Study session checks the user's ban state. If an active ban exists, the join is rejected and the ban reason and expiry are shown.
Alternate/Exceptional Flow:
E-1 Room ended between the list being displayed and the join being submitted. System shows that the room is no longer available and refreshes the list.
E-2 Room is at or above capacity, including the case where two users contend for the final slot. System rejects the join atomically, shows that the room is full, and refreshes the list.
E-3 User is already an active participant in a room. System offers to leave that room (UC-03) or cancel.
E-4 Missing or expired JWT. Gateway rejects the request and redirects to Sign In.
E-5 User has an active ban. Join is denied.
E-6 Real-time connection drops after the participant record is written. The record persists and the client re-syncs presence on reconnect without creating a duplicate row.

Use Case Name:
Leave Study Session Room
ID: UC-03
Important Level: High
Primary Actor: Customer (authenticated user)
Use Case Type: Detail, Essential
Stakeholders and Interests:
Customer : wants to exit cleanly, keep every reward already earned during the session, and see a summary of what they completed.
Admin : wants participant lists to stay accurate and empty rooms removed from monitoring.
Brief Description:
A participant leaves the room. Their timer is stopped and finalized and their participation record is closed. No reward is granted here.
Trigger: User selects "Leave Room".
Type: External
Relationships:
Owned by Study session service. Triggers timer finalization in Study timer. Publishes participant.left to Admin monitoring.
Precondition:
User is signed in; gateway has validated a non-expired JWT.
User is an active participant in a StudySession with status = ACTIVE.
User's timer state for that room is known: running, paused, resting, or stopped.
Postcondition:
SessionParticipant closed with left_at set.
User's TimerSession for the room is finalized; no running timer is left orphaned.
participant.left published; Admin monitoring updated.
If no participants remain, StudySession.status = ENDED and session.ended is published.
Normal Flow of Events:
User selects Leave Room.
System warns if a work cycle is currently in progress (S-1).
System stops and finalizes the user's timer for this room, discarding the in-progress cycle.
System sets left_at on the SessionParticipant record.
System publishes participant.left.
System updates the live participant list for remaining members and for Admin monitoring.
System displays a session summary showing completed cycles, total focus time, and the rewards already earned during this session.
System returns the user to the room list.
If the user was the last participant, the system ends the room (S-2).
Subflows:
S-1 Confirm leave during an active cycle. System states that the in-progress cycle will not be completed and will not award a reward. User confirms or cancels.
S-2 End empty room. When the participant count reaches zero, status is set to ENDED, closed_at is stamped, and the room is removed from the public room list and from Admin monitoring.
Alternate/Exceptional Flow:
E-1 User cancels at the confirmation prompt. The user remains in the room and the timer continues untouched.
E-2 User completed no cycles. Participation closes normally and the summary reports no rewards earned.
E-3 Connection is lost without an explicit leave. After a timeout the system auto-leaves the user and finalizes the timer at the last completed cycle boundary. Rewards from completed cycles are already banked and are unaffected.
E-4 Admin kicks the participant. The flow runs from step 3 as an alternate entry point. The kick action itself belongs to AdminModeration.
E-5 Persistence failure while closing the participation. The timer stays finalized, the leave is retried, and the operation is idempotent on (session_id, user_id).

Use Case Name: Admin Monitor Active Study Sessions
ID: UC-04
Important Level: High
Primary Actor: Admin
Use Case Type: Detail, Essential
Stakeholders and Interests:
Admin : wants a live view of every active room and its participants to spot abuse.
Customer : wants monitoring to never affect their own timer.
Brief Description: An admin views all active rooms with their participants and timer states, updated in real time.
Trigger: Admin opens the monitoring dashboard.
Type: External
Relationships:
Owned by Admin service.
Subscribes independently to Study session and Study timer via Realtime, joined client-side by session_id and user_id.
Extends to AdminModeration for kick and ban.
Precondition:
Admin is signed in; gateway has validated a non-expired JWT with role = ADMIN.
Realtime channels are reachable.
Postcondition:
No session or timer state is modified.
Dashboard reflects the current live state.
Access is written to the admin audit log.
Normal Flow of Events:
Admin opens the monitoring dashboard.
The system verifies the role is ADMIN (E-1).
System loads all StudySession records with status = ACTIVE.
The system loads the TimerSession records of their participants.
The system joins both sets by session_id and user_id and lists the rooms with name, participant count, and capacity.
The system subscribes to both Realtime channels (E-2, E-3).
Admin selects a room (S-1).
The system shows each participant's timer state, remaining time, and completed cycles.
The system applies incoming events to the view without a reload (E-4).
Admin may escalate a participant to moderation (S-2).
Admin closes the dashboard and both subscriptions are released.
Subflows:
S-1 Inspect room: System opens the room detail panel and streams that room's updates.
S-2 Escalate: Admin selects a participant and hands off to AdminModeration. The action itself is out of scope.
Alternate/Exceptional Flow:
E-1 Caller is not an admin. Request is rejected and no session data is exposed.
E-2 Realtime connection drops. View is marked stale and a fresh snapshot is loaded on reconnect, not a replay.
E-3 Timer data is unavailable. Rooms and participants are still listed, with timer state shown as unknown.
E-4 A room ends while being inspected. It is marked ended and the admin returns to the list.
E-5 Missing or expired JWT. Gateway rejects the request and redirects to Sign In.

Use Case Name: Manage Study Timer
ID: UC-05
Important Level: High
Primary Actor: Customer (participant in a room)
Use Case Type: Detail, Essential
Stakeholders and Interests:
Customer : wants full control of a personal timer, and confidence that a completed cycle is banked.
Reward : wants a reliable completion signal carrying cycle length and active timer count.
Admin : wants each timer state to be observable.
Brief Description: A participant configures and operates their own timer inside a room, keyed by (session_id, user_id) and never synchronized with others. Completing a work cycle emits the event that drives UC-09, stopping or abandoning one does not.
Trigger: User operates a timer control.
Type: External
Relationships:
Owned by Study timer service. Owns TimerSession, TimerCycle, TimerSetting.
Publishes cycle.completed to Reward.
Timer state is read by Admin monitoring via Realtime.
Precondition:
User is signed in and gateway has validated a non-expired JWT.
User is an active participant in a StudySession with status = ACTIVE.
A TimerSetting exists, or the default applies.
Postcondition:
TimerSession reflects the resulting state, with one TimerCycle per started cycle marked completed, skipped, or discarded.
cycle.completed is published for every completed work cycle with cycle_length and active_timer_count.
Updated state is visible to the room and to Admin monitoring.
Normal Flow of Events:
User opens the timer panel; system loads the saved TimerSetting (S-1).
User selects Start.
System verifies the user is still an active participant (E-2).
System verifies no cycle is already running (E-4).
System opens a TimerCycle with type = WORK, recording started_at, duration, and active_timer_count.
System broadcasts RUNNING to the room and to Admin monitoring.
User may pause and resume during the cycle (S-2).
When the duration elapses, system marks the cycle completed and publishes cycle.completed.
System starts the rest period (S-3).
When rest ends or is skipped, the timer returns to a ready state for the next cycle.
User may stop the timer at any point (S-4).
System updates the participant's completed cycles and total focus time.
Subflows:
S-1 Update setting: User sets the work and rest durations within the permitted range (E-1), defaulting to 20 minutes of work and 5 minutes of rest. One cycle runs work, then rest, then returns to a ready state, and repeats until the user stops. Values are persisted and take effect from the next cycle.
S-2 Pause and resume: System freezes the elapsed time and broadcasts PAUSED, then resumes from the remaining time. The cycle is not forfeited (E-3).
S-3 Rest period: System opens a REST cycle after a completed work cycle. The user may skip it. No reward is rolled for rest.
S-4 Stop: System ends the current cycle. A work cycle in progress is discarded, no event is published, and its rewards are forfeited. Completed cycles are unaffected.
Alternate/Exceptional Flow:
E-1 Duration is outside the permitted range. Validation error is shown and the timer is not started.
E-2 User is no longer an active participant, having left, been kicked, or the room ended. The cycle in progress is discarded and the timer is finalized.
E-3 A cycle stays paused beyond the permitted limit. System stops the timer and discards the cycle, so an idle pause cannot bank a cycle.
E-4 Duplicate Start requests. System keeps a single active cycle per (session_id, user_id) and ignores the duplicate.
E-5 Client disconnects or reloads. Progress is derived from server-recorded timestamps, so the timer is unaffected and the client re-syncs on reconnect.
E-6 Missing or expired JWT. Gateway rejects the request and redirects to Sign In.

Use Case Name: Auth (SignIn & SignUp)
ID: UC-06
Important Level: High
Primary Actor: Unauthenticated User
Use Case Type: Detail, Essential
Stakeholders and Interests:
Customer : wants a fast and easy way to sign in/up.
Brief Description: An unauthenticated User signs in or signs up using Google OAuth. The system exchanges the OAuth code for the user's Google profile, creates a new User record on first login or matches an existing one by email, and issues a JWT carrying the user's email and role. No separate password is collected at this step.
Trigger: User selects "Sign in with Google".
Type: External
Relationships:
Owned by Auth service.
Precondition:
Google OAuth endpoint is reachable.
The user has a valid Google account.
Postcondition:
An User record exists with email and role set (created if this is the first login).
A JWT is issued containing user_id, email, and role.
An authenticated session is established.
Normal Flow of Events:
The user selects "Sign in with Google".
The system redirects the user to the Google OAuth consent screen.
The user authorizes the app.
Google redirects back with an authorization code.
The system exchanges the code for the user's Google profile (email, display name).
The system checks whether a User with that email already exists (S-1).
System issues a JWT containing user_id, email, and role.
The system establishes the session and redirects the user to the room list.
Subflows:
S-1 Match or create account: If a User with the returned email exists, the system loads their stored role and signs them in. If none exists, the system creates a new User record with that email, the display name from the Google profile, and role = CUSTOMER by default.
Alternate/Exceptional Flow:
E-1 User denies consent on the Google screen. System returns to the sign-in page; no session or record is created.
E-2 Authorization code exchange fails or the code has expired. System shows an error and lets the user retry.
E-3 Persistence failure while creating a new User. No partial record is retained; an error is shown and the user may retry.
E-4 Email belongs to an account with an active ban. Sign-in still succeeds and a session is issued; the ban itself is enforced later, at room join (UC-02), not at login.
E-5 Role = ADMIN is never assigned through this flow. Admin accounts are pre-provisioned directly in the User table; a normal Google sign-in can only ever result in role = CUSTOMER.

Use Case Name: Manage Account
ID: UC-07
Important Level: Medium
Primary Actor: Customer (authenticated user)
Use Case Type: Detail, Essential
Stakeholders and Interests:
Customer : wants to keep their display name current, and wants a single place to see how their studying is going.
Brief Description: An authenticated user updates their display name, and views a personal dashboard summarizing their timer activity: total sessions, total focus time, and rewards earned. The dashboard is read-only and lives on the account page.
Trigger: User opens the Account page.
Type: External
Relationships:
Owned by Account management service.
Reads aggregated data from Study timer and Reward.
Publishes profile.updated to Study session and Leaderboard, so display name changes propagate.
Precondition:
User is signed in and gateway has validated a non-expired JWT.
Postcondition:
If a name change was submitted: User.display_name updated and profile.updated published.
Dashboard reflects the user's current timer and reward statistics at the time of viewing.
Normal Flow of Events:
The user opens the Account page.
The system loads the current profile (display name, email) and the personal dashboard (S-1).
Users edit their display name and submit (S-2).
The system confirms the update and refreshes the displayed profile.
Subflows:
S-1 Build dashboard summary: System aggregates the user's TimerSession and TimerCycle records for total sessions joined, and total focus time, and aggregates UserReward for rewards earned, then renders all four on the account page.
S-2 Update display name: System validates the new name (E-1), updates User.display_name, and publishes profile.updated so it is reflected in active rooms and on the leaderboard without requiring re-login.
Alternate/Exceptional Flow:
E-1 New display name is empty or exceeds the max length. System shows a validation error and keeps the previous name.
E-2 Persistence failure while saving the profile change. No partial update is retained; an error is shown and the user may retry.
E-3 Missing or expired JWT. Gateway rejects the request and redirects to Sign In.

Use Case Name: View Leaderboard
ID: UC-08
Important Level: Medium
Primary Actor: Customer (authenticated user)
Use Case Type: Detail, Essential
Stakeholders and Interests:
Customer : wants to see how their progress compares to other users, as motivation to keep studying.
Brief Description: A customer views the FishTank leaderboard, which ranks users by rewards earned. The user can filter the ranking by weekly, monthly, or all-time periods. Rankings are read-only.
Trigger: User opens the Leaderboard.
Type: External
Relationships:
Owned by Leaderboard service.
Consumes reward.awarded from Reward to keep totals current.
Reads User for display name.
Precondition:
User is signed in and gateway has validated a non-expired JWT.
At least one reward has been awarded, or the leaderboard is shown empty.
Postcondition:
No ranking or reward data is modified.
The requested period's ranking is displayed to the user.
Normal Flow of Events:
User selects Leaderboard.
The system defaults to the all-time period and loads that ranking (S-1).
The system displays users ranked by total rewards, with display name and reward count/value.
Users select a different period: Weekly, Monthly, or All-Time (S-2).
System reloads and displays the ranking for the selected period.
Users may locate their own position, highlighted in the list.
Subflows:
S-1 Compute ranking: System sums UserReward value per user, scoped to the selected period's date range (or unbounded for all-time), orders descending, and assigns rank positions.
S-2 Filter by period: System re-runs S-1 with the newly selected period's date range and re-renders the list without a full page reload.
Alternate/Exceptional Flow:
E-1 No rewards exist yet for the selected period. System shows an empty state rather than an error.
E-2 User has never earned a reward. User still sees the full ranking but no highlighted position for themselves.
E-3 Missing or expired JWT. Gateway rejects the request and redirects to Sign In.
E-4 Tie in total rewards between two or more users. System breaks ties by earliest reward timestamp, so the user who reached that total first ranks higher.

Use Case Name:
Award Reward on Cycle Completion
ID: UC-09
Important Level: High
Primary Actor: System (Reward service)
Use Case Type: Detail, Essential
Stakeholders and Interests:
Customer : wants rewards for finishing a work cycle, and better odds when studying alongside others.
Admin : wants the community feature to be mechanically worth using, and wants abandoning a cycle to carry a visible cost.
Brief Description:
When a participant completes a work cycle, the system rolls a set of rewards for them. The number of rewards scales with cycle length, and rarity is weighted by how many participants were actively running a timer in the same room when the cycle began. Rewards accumulate visibly during the cycle but are only banked on completion; abandoning a cycle forfeits all of them.
Trigger: A work cycle completes in Study timer.
Type: System-triggered
Relationships:
Owned by Reward service.
Consumes cycle.completed from Study timer.
Publishes reward.awarded to Leaderboard.
Precondition:
A work cycle, not a break cycle, has been completed for a participant.
A RewardItem catalogue with rarity weights exists.
Postcondition:
A UserReward record created for the completed cycle, linked to the user and to cycle_id, holding the awarded items.
reward.awarded published; leaderboard totals updated.
Rewards are visible in the user's collection and personal dashboard.
Normal Flow of Events:
A participant's work cycle completes in the Study timer.
Study timer counts participants actively running a timer in that room, snapshotted at cycle start.
Study timer publishes cycle.completed carrying user_id, session_id, cycle_id, cycle_length, and active_timer_count.
Reward consumes the event.
Reward verifies that no reward already exists for this cycle_id (E-1).
Reward determines the number of items to award, scaled by cycle_length (S-1).
Reward applies the community buff to the rarity weights based on active_timer_count (S-2).
Reward rolls the items from RewardItem against the adjusted weights.
Reward creates the UserReward record holding the rolled items.
Reward publishes reward.awarded.
Leaderboard consumes the event and updates weekly, monthly, and all-time totals.
Client reveals the awarded items to the user.
Subflows:
S-1 Determine award count. A longer work cycle yields more items, so a participant running long focus blocks is rewarded proportionally to time studied.
S-2 Apply community buff. A higher active_timer_count shifts the drop table toward rarer items. A count of 1, meaning solo study, uses the base table.
Alternate/Exceptional Flow:
E-1 A reward already exists for this cycle_id, caused by duplicate event delivery. The event is acknowledged and no second reward is created.
E-2 Reward service is unavailable. The event is retained by the broker and replayed on recovery. Rewards arrive late but are not lost.
E-3 A break cycle completes. No event is published and no reward is rolled.
E-4 Participant abandons the cycle before it completes, by stopping the timer or leaving the room. No cycle.completed is published, so nothing is banked and all accumulated rewards are forfeited.
E-5 Leaderboard is unavailable. reward.awarded is retained and replayed. Rewards are already banked; only the ranking lags.

Functional Requirements
Study Session Room System
The system shall allow an authenticated customer to create a study session room with a room name and a participant limit.
The system shall allow a participant limit between 1 and a configured system maximum, where a limit of 1 produces a solo session that no other user can join.
The system shall enroll the room creator as the first participant of the room.
The system shall publish every created room to the public room list.
The system shall allow customers to browse active rooms with their name, current participant count, and capacity.
The system shall allow an authenticated customer to join an active room that has remaining capacity.
The system shall reject a join request when the room has ended or has reached its participant limit.
The system shall display the live participant list to all members of a room in real time.
The system shall allow a participant to leave a room at any time.
The system shall stop and finalize a participant's timer when that participant leaves the room.
The system shall warn a participant that an in-progress work cycle will be forfeited before confirming a leave.
The system shall display a session summary on leave showing completed cycles, total focus time, and rewards earned during the session.
The system shall end a study session room automatically when its last participant leaves.
The system shall prevent a user from being an active participant in more than one study session room at a time.
The system shall maintain one independent timer per participant per room.
Admin Monitoring System
The system shall restrict all monitoring functions to users holding the administrator role.
The system shall allow an administrator to view all study session rooms with status ACTIVE, with their name, participant count, and capacity.
The system shall display each participant's timer state, remaining time, and completed cycle count in a selected room.
The system shall update monitored information in real time without requiring a reload.
The system shall prevent administrators from modifying any participant's timer.
The system shall display session information when timer information is unavailable, and shall indicate when data is no longer live.
Study Timer System
The system shall maintain one independent timer per participant per room, identified by session and user.
The system shall allow a participant to configure work and rest durations within a system-defined range, and shall persist them for later sessions.
The system shall apply a default of 20 minutes of work and 5 minutes of rest to a participant who has not configured a timer setting.
The system shall apply changed settings from the next cycle only.
The system shall record the start time, duration, and number of participants actively running a timer in the room at the start of each work cycle.
The system shall allow a participant to pause and resume a work cycle without forfeiting it, and shall discard a cycle paused beyond a system-defined limit.
The system shall allow a participant to stop a timer at any time, discarding the cycle in progress without awarding rewards.
The system shall complete a work cycle automatically when its duration has elapsed, and publish a completion event carrying the cycle length and active timer count.
The system shall start a rest period after a completed work cycle, allow it to be skipped, and award no reward for it.
The system shall maintain a single active cycle per participant per room.
The system shall derive timer progress from server-recorded timestamps so that disconnection or reload does not alter the timer.
The system shall broadcast each participant's timer state to the room and to administrative monitoring in real time.
The system shall finalize a participant's timer when they leave the room or the room ends.
The system shall record completed work cycles and total focus time for the participant's dashboard.
Authentication System
The system shall allow a visitor to sign in or sign up exclusively via Google OAuth.
The system shall create a new User record on first login, populated with email and display name from the Google profile.
The system shall match an existing User by email on subsequent logins rather than creating a duplicate.
The system shall assign role = CUSTOMER by default to every newly created account.
The system shall restrict role = ADMIN to accounts provisioned directly, never through Google sign-in.
The system shall issue a JWT containing user_id, email, and role upon successful authentication.
The system shall allow an authenticated user to sign out, invalidating their current session.
The system shall reject requests bearing a missing or expired JWT and redirect the user to Sign In.
Account Management System
The system shall allow an authenticated user to view their profile: display name and email.
The system shall allow an authenticated user to update their display name.
The system shall propagate a display name change to active rooms and the leaderboard without requiring re-login.
The system shall display a personal dashboard summarizing total sessions joined, cycles completed, total focus time, and rewards earned.
The system shall keep the dashboard read-only, deriving all values from existing timer and reward records rather than storing separate copies.
Leaderboard System
The system shall rank users by total rewards earned, filterable by weekly, monthly, and all-time periods.
The system shall default to the all-time period when the leaderboard is first opened.
The system shall recompute a period's ranking from UserReward records scoped to that period's date range.
The system shall update leaderboard totals upon receiving a reward.awarded event.
The system shall highlight the signed-in user's own position within the displayed ranking, when present.
The system shall break ranking ties by earliest reward timestamp.
The system shall display the leaderboard with the last-known ranking, with a "last updated" indicator, when reward.awarded delivery is delayed.
Reward System
The system shall award rewards to a participant upon completion of a work cycle.
The system shall forfeit all accumulated rewards when a participant abandons a work cycle before completion.
The system shall scale the number of rewards awarded with the length of the completed work cycle.
The system shall weight reward rarity by the number of participants actively running a timer in the same room at the start of the cycle.
The system shall display a participant's awarded rewards in their collection and personal dashboard.

Non-Functional Requirements
Operational
The system shall support access through major modern web browsers on desktop and mobile devices.
The system shall maintain study session, timer, reward, and leaderboard data after users leave or re-enter the system.
The system shall recover the latest saved study session state after an unexpected service interruption.
Performance
The system shall respond to common user requests within 2 seconds under normal operating conditions.
The system shall synchronize study session status and participant information among active users within 2 seconds under normal network conditions.
The system shall maintain acceptable response time when supporting at least 50 concurrent users.
Security
The system shall authenticate users through Google OAuth.
The system shall enforce role-based access control between customers and administrators.
The system shall prevent users from accessing or modifying other users' personal information and study statistics.
The system shall restrict administrative functions, such as monitoring, reporting, and banning users, to authorized administrators.
The system shall protect users' personal information in accordance with applicable PDPA requirements.
Cultural and Legal
The system shall provide Thai as the default language for the user interface.
The system shall comply with applicable Thailand's Personal Data Protection Act (PDPA) requirements regarding users' personal information.
Usability
The system shall provide a responsive user interface for desktop and mobile devices.
The system shall allow users to start, pause, resume, and stop their study timers with minimal interaction.
The system shall provide clear visual feedback for work and rest periods.
The system shall provide intuitive navigation between StudySession, StudyTimer, Dashboard, Reward, and Leaderboard features.
The system shall minimize the number of steps required to create or join a StudySession.

Architecture Decision Records (ADRs)
ADR-001: Frontend Framework

Context
The system requires a responsive user interface that managing real-time states like study timers while maintaining fast initial load times and good SEO for the study session pages.

Options
React: Standard SPA. Pro: Massive ecosystem. Con: Lacks built-in routing and SSR.
Vue.js: Lightweight framework. Pro: Clean syntax. Con: Smaller ecosystem for timer/state libraries.
Next.js: React framework with SSR. Pro: Built-in routing and fast loads. Con: Harder server/client boundary.

Decision
Use Next.js (React) as the frontend framework.

Status
Accepted

Consequences
Requires to separate interactive Client components from Server components.
Potential for hydration mismatches if the live timer state isn't managed perfectly between the server and browser.

ADR-002: CSS/UI Framework

Context
The application features a unique visual theme (Fishing/FishTank gamification) that must be integrated alongside standard UI elements like timer settings and admin dashboards.

Options
Bootstrap: Grid-based framework. Pro: Quick setup. Con: Outdated style.
Material UI (MUI): Component library. Pro: Ready-to-use. Con: Rigid design system fights custom themes.
TailwindCSS: Utility-first CSS. Pro: Total design freedom for the FishTank theme. Con: Can clutter HTML.

Decision
Adopt TailwindCSS for UI styling.

Status
Accepted

Consequences
Files can easily become messy and hard to read due to long strings of utility classes.
Forces the team to strictly build and reuse Next.js components to keep code clean.

ADR-003: Backend Framework

Context
The backend must handle high concurrency for real-time community study sessions, active timer synchronizations, and leaderboard updates while keeping the codebase maintainable.

Options
Go (Golang): Compiled language. Pro: Great concurrency, supports Hexagonal/DDD patterns. Con: Steeper learning curve.
Node.js (NestJS): JS backend. Pro: Same language as frontend. Con: Struggles with heavy concurrent tasks.
Python (Django): Feature-rich framework. Pro: Fast development. Con: Lower performance for real-time web.

Decision
Use Go (Golang) for backend API and session state management.

Status
Accepted

Consequences
Steeper learning curve for the team to master Go’s strict syntax and error handling
Developers must be careful when managing concurrent tasks (goroutines) to avoid memory leaks or race conditions.
ADR-004: Database Strategy

Context
The system has two distinct data requirements: real-time speed for active timers/sessions, and flexible, nested data structures for gamification items and rankings.

Options
PostgreSQL : Relational DB. Pro: Strong data integrity. Con: Rigid schema for flexible JSON.
MongoDB: NoSQL DB. Pro: Perfect for nested object structures. Con: Lacks built-in WebSockets.
Supabase + MongoDB Pro: Supabase WebSockets for speed, Mongo for objects. Con: Harder to maintain.

Decision
Adopt a polyglot approach using Supabase and MongoDB.

Status
Accepted

Consequences
High architectural complexity because the backend has to maintain connections and queries for two totally different databases.
Difficult to ensure data consistency across both databases. For example, if a user gets deleted in Supabase, we have to make sure their data is deleted in MongoDB.

ADR-005: Authentication Strategy

Context
To encourage platform adoption, the barrier to entry must be as low as possible. Managing passwords introduces friction and security overhead.

Options

Email & Password: Pro: Full control over user data. Con: High friction and requires password reset flows.
SMS OTP: Pro: Highly secure and verifies real users easily. Con: Incurs ongoing operational costs.
Google OAuth: Pro: Seamless one-click login for students. Con: Excludes users without Google accounts.

Decision
Use Google OAuth as the primary authentication mechanism for Sign Up and Sign In.

Status
Accepted

Consequences
Completely locks out users who do not have a Google account.
Creates a strict dependency on Google's external servers

ADR-006: Deployment Infrastructure

Context
The infrastructure needs to host both a Next.js frontend and a Go backend with minimal DevOps overhead, allowing the team to focus on feature development.

Options
AWS: Industry standard. Pro: Maximum control and scalability. Con: High DevOps overhead and complex setup.
Vercel + Heroku: Split hosting. Pro: Optimized for each stack. Con: Fragmented monitoring across dashboards.
Render: Unified hosting. Pro: Hosts Next.js and Go easily from GitHub. Con: Less granular infrastructure control.

Decision
Deploy both the Frontend (Next.js) and Backend (GoLang) on Render.

Status
Accepted

Consequences
Lower-tier or free plans might suffer from "cold starts," causing the app to load slowly for the first user after a period of inactivity.
Provides much less control over server configurations, custom networking, and firewalls compared to AWS.
