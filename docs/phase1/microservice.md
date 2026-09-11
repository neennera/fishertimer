| Service       | Operations                                                                                                                                                                    | Collaborators                                                            |
| ------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------ |
| Auth          | AuthenticateUser()<br>SignUp()<br>SignOut()                                                                                                                                   | -                                                                        |
| Account       | ViewProfile()<br>UpdateProfile()<br>ChangePassword()<br>ViewStatistics()<br>UpdateBanStatus()                                                                                 | StudyTimer Service TimerStatistics()<br>Reward Service ViewRewards()     |
| Study Session | CreateSession()<br>JoinSession()<br>LeaveSession()<br>EndSession()<br>SetParticipantLimit()                                                                                   | Admin Service VerifyBanStatus()                                          |
| Study Timer   | StartTimer()<br>PauseTimer()<br>ResumeTimer()<br>StopTimer()<br>StartRest()<br>SkipRest()<br>CompleteCycle()<br>UpdateTimerSetting()<br>ResetTimer()<br>TimerStatistics()     | -                                                                        |
| Reward        | AwardReward()<br>ClaimReward()<br>ViewRewards()<br>TrackProgression()                                                                                                         | -                                                                        |
| Leaderboard   | ViewLeaderboard()<br>GetRanking()<br>FilterByPeriod()                                                                                                                         | -                                                                        |
| Admin         | ViewActiveSessions()<br>ViewParticipants()<br>MonitorTimerStatus()<br>ViewSessionDetails()<br>ReportUser()<br>ReviewReport()<br>BanUser()<br>UnbanUser()<br>VerifyBanStatus() | StudySession Service LeaveSession()<br>Account Service UpdateBanStatus() |

Auth
Detail : Manages Google OAuth authentication and creates user records on first login.
Operations :
AuthenticateUser : Authenticates users via Google OAuth.
SignUp : Creates a new user profile on first login.
SignOut : Invalidates the user's current session.
Collaboration :
None

Account
Detail : Handles user profiles, ban states, and renders the personal dashboard.
Operations :
ViewProfile : Displays the user's current email, display name, and ban status.
UpdateProfile : Updates user details like display name and ban flags.
ChangePassword : Manages user password changes.
ViewStatistics : Aggregates total sessions, focus time, and rewards earned.
UpdateBanStatus : Updates the user's ban flag in the database.
Collaboration :
StudyTimer Service TimerStatistics() : Fetches total sessions and focus time for the dashboard.
Reward Service ViewRewards() : Fetches earned rewards for the dashboard.

Study Session
Detail : Manages the lifecycle of study rooms and participant limits.
Operations :
CreateSession : Opens a new study room with configured capacities.
JoinSession : Adds a participant to the active room roster.
LeaveSession : Removes a participant from the active room roster.
EndSession : Closes a room when the last participant leaves.
SetParticipantLimit : Sets the maximum number of users allowed in a session.
Collaboration :
Admin Service VerifyBanStatus() : Checks the user's ban state before allowing room creation or joins.

Study Timer
Detail : Runs independent user timers within a session and handles work and rest cycles.
Operations :
StartTimer : Initiates a new timer cycle.
PauseTimer : Pauses an ongoing work cycle without forfeiting it.
ResumeTimer : Resumes a previously paused work cycle.
StopTimer : Finalizes a timer and discards in-progress cycles.
StartRest : Manages the break period between work cycles.
SkipRest : Bypasses the rest period to return to a ready state.
CompleteCycle : Finishes a work duration and publishes completion data.
UpdateTimerSetting : Modifies the default work and rest durations.
ResetTimer : Resets the timer progress.
TimerStatistics: Exposes timer session data for the user dashboard.
Collaboration :
None

Reward
Detail : Calculates and grants items based on cycle length and active participant count.
Operations :
AwardReward : Rolls for items using a rarity table buffed by community presence.
ClaimReward : Processes the user claiming a granted reward.
ViewRewards : Displays accumulated items in the user's collection.
TrackProgression : Monitors user milestones.
Collaboration :
None

Leaderboard
Detail : Provides a read-optimized ranking of users based on total rewards.
Operations :
ViewLeaderboard : Displays the leaderboard dashboard interface.
GetRanking : Retrieves the list of users ranked by total rewards.
FilterByPeriod : Scopes rankings to Weekly, Monthly, or All-Time.
Collaboration :
None

Admin
Detail : Bundles real-time session monitoring with moderation capabilities.
Operations :
ViewActiveSessions : Subscribes to live room states.
ViewParticipants : Lists the users currently in an active session.
MonitorTimerStatus : Subscribes to live timer states for participants.
ViewSessionDetails : Displays comprehensive information about a specific room.
ReportUser : Submits a moderation report against a user.
ReviewReport : Evaluates a submitted user report.
BanUser : Applies platform restrictions on users.
UnbanUser : Removes platform restrictions from users.
VerifyBanStatus : Returns the ban status of a user to other services.
Collaboration :
StudySession Service LeaveSession() : Sends a command to kick users from active rooms.
Account Service UpdateBanStatus() : Sends a command to set or remove the ban flag on the user account.

Diagram
