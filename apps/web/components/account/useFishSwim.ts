"use client";

import { useEffect, type RefObject } from "react";

// Units are art pixels (var(--px)), unrounded: tank fish are the documented
// exception to the whole-pixel sprite rule.
const DEFAULT_FISH_SIZE = 16;
const EDGE = 2;
// Keeps fish out of the sand and rocks.
const FLOOR_FRACTION = 0.16;

// Art px per second.
const CRUISE_MIN = 7;
const CRUISE_MAX = 13;
// Lower = wider turns.
const STEER = 1.4;
const DART_CHANCE = 0.1; // per second
const DART_SPEED = 2.4; // x cruise
const DART_SECONDS = 0.6;
const IDLE_CHANCE = 0.08; // per second
const IDLE_SPEED = 0.12; // x cruise
const IDLE_SECONDS: [number, number] = [1.2, 2.8];
// Hysteresis, so a fish drifting vertically doesn't keep flipping.
const TURN_THRESHOLD = 1.2;
const TURN_RATE = 7;
// Degrees.
const MAX_TILT = 28;
const TILT_RATE = 5;
const SWAY = 3;
// Depth: 0 = back, 1 = front.
const DEPTH_SCALE: [number, number] = [0.62, 1];
const DEPTH_OPACITY: [number, number] = [0.7, 1];
const DEPTH_SPEED: [number, number] = [0.6, 1];
const DEPTH_RATE = 0.12; // depth units per second
const PERSONAL_SPACE = 20;
const SPACE_DEPTH = 0.5;
const SPACE_PUSH = 9;
// Random goals across the whole tank make fish bunch in the middle, so each
// mostly roams around its home spot and picks the least crowded candidate.
const HOME_RANGE = 0.3; // 0–1 of the tank
const ROAM_CHANCE = 0.2;
const GOAL_CANDIDATES = 4;
// Tail beats per second.
const TAIL_IDLE = 2;
const TAIL_PER_SPEED = 0.35;
const TAIL_FRAMES = [0, 1, 0, 2];
// So a tab coming back doesn't teleport the fish.
const MAX_DT = 0.1;

/** Rest position (and home), 0–1 of the space the fish can swim in. */
export interface SwimStart {
  restX: number;
  restY: number;
}

interface Swimmer {
  /** Hidden on phones; ignored when others look for room. */
  shown: boolean;
  homeX: number;
  homeY: number;
  x: number;
  y: number;
  z: number;
  vx: number;
  vy: number;
  goalX: number;
  goalY: number;
  goalZ: number;
  goalTimer: number;
  cruise: number;
  mode: "cruise" | "dart" | "idle";
  modeTimer: number;
  facing: 1 | -1;
  /** -1..1, eased toward `facing`: the turn narrows through side-on. */
  turn: number;
  tilt: number;
  tailPhase: number;
  bobPhase: number;
}

function between(min: number, max: number) {
  return min + Math.random() * (max - min);
}

function mix([from, to]: [number, number], t: number) {
  return from + (to - from) * t;
}

function approach(value: number, target: number, rate: number, dt: number) {
  return value + (target - value) * Math.min(1, rate * dt);
}

/**
 * Animates the `[data-fish]` elements in the tank. Writes styles directly, so
 * React doesn't re-render per frame. Off under prefers-reduced-motion (fish
 * stay at rest) and paused while the tank is off-screen.
 */
export function useFishSwim(waterRef: RefObject<HTMLElement | null>, starts: SwimStart[]) {
  const startsKey = JSON.stringify(starts);

  useEffect(() => {
    const water = waterRef.current;
    if (!water) {
      return;
    }
    const fishEls = Array.from(water.querySelectorAll<HTMLElement>("[data-fish]"));
    const parts = fishEls.map((el) => ({
      turn: el.querySelector<HTMLElement>(".pixel-fish__turn"),
      frame: el.querySelector<HTMLElement>("[data-frame]"),
    }));
    const reducedMotion = window.matchMedia("(prefers-reduced-motion: reduce)");

    const artPx = parseFloat(getComputedStyle(water).getPropertyValue("--px")) || 1;
    let width = 0;
    let height = 0;
    let fishSize = DEFAULT_FISH_SIZE;
    const swimmers: Swimmer[] = [];
    // Re-read on resize: the phone breakpoint changes fish size and count.
    function measure() {
      width = water!.clientWidth / artPx;
      height = water!.clientHeight / artPx;
      fishSize =
        parseFloat(getComputedStyle(water!).getPropertyValue("--fish-size")) || DEFAULT_FISH_SIZE;
      swimmers.forEach((fish, i) => {
        fish.shown = getComputedStyle(fishEls[i]!).display !== "none";
      });
    }
    measure();

    const maxX = () => Math.max(EDGE, width - fishSize - EDGE);
    const maxY = () => Math.max(EDGE, height * (1 - FLOOR_FRACTION) - fishSize);

    function candidateGoal(fish: Swimmer): [number, number] {
      const roam = Math.random() < ROAM_CHANCE;
      const fx = roam ? Math.random() : fish.homeX + between(-HOME_RANGE, HOME_RANGE);
      const fy = roam ? Math.random() : fish.homeY + between(-HOME_RANGE, HOME_RANGE);
      const clamp = (v: number) => Math.min(1, Math.max(0, v));
      return [EDGE + clamp(fx) * (maxX() - EDGE), EDGE + clamp(fy) * (maxY() - EDGE)];
    }

    function pickGoal(fish: Swimmer) {
      let best: [number, number] = [fish.x, fish.y];
      let bestRoom = -1;
      for (let c = 0; c < GOAL_CANDIDATES; c++) {
        const [gx, gy] = candidateGoal(fish);
        let room = Infinity;
        for (const other of swimmers) {
          if (other !== fish && other.shown) {
            room = Math.min(
              room,
              Math.hypot(other.x - gx, other.y - gy),
              Math.hypot(other.goalX - gx, other.goalY - gy),
            );
          }
        }
        if (room > bestRoom) {
          bestRoom = room;
          best = [gx, gy];
        }
      }
      [fish.goalX, fish.goalY] = best;
      fish.goalZ = Math.random();
      fish.goalTimer = between(4, 9);
    }

    const parsedStarts = JSON.parse(startsKey) as SwimStart[];
    fishEls.forEach((_, i) => {
      const start = parsedStarts[i] ?? { restX: 0.5, restY: 0.5 };
      const facing = Math.random() < 0.5 ? 1 : -1;
      const fish: Swimmer = {
        shown: true,
        homeX: start.restX,
        homeY: start.restY,
        x: start.restX * maxX(),
        y: EDGE + start.restY * (maxY() - EDGE),
        z: 1,
        vx: 0,
        vy: 0,
        goalX: 0,
        goalY: 0,
        goalZ: 1,
        goalTimer: 0,
        cruise: between(CRUISE_MIN, CRUISE_MAX),
        mode: "cruise",
        modeTimer: 0,
        facing,
        turn: facing,
        tilt: 0,
        tailPhase: Math.random() * TAIL_FRAMES.length,
        bobPhase: Math.random() * Math.PI * 2,
      };
      fish.goalX = fish.x;
      fish.goalY = fish.y;
      swimmers.push(fish);
    });
    measure();
    swimmers.forEach(pickGoal);

    function step(dt: number) {
      for (const fish of swimmers) {
        fish.modeTimer -= dt;
        if (fish.mode !== "cruise" && fish.modeTimer <= 0) {
          fish.mode = "cruise";
        } else if (fish.mode === "cruise") {
          if (Math.random() < DART_CHANCE * dt) {
            fish.mode = "dart";
            fish.modeTimer = DART_SECONDS;
            pickGoal(fish);
          } else if (Math.random() < IDLE_CHANCE * dt) {
            fish.mode = "idle";
            fish.modeTimer = between(...IDLE_SECONDS);
          }
        }

        fish.goalTimer -= dt;
        const toX = fish.goalX - fish.x;
        const toY = fish.goalY - fish.y;
        const distance = Math.hypot(toX, toY);
        if (distance < 3 || fish.goalTimer <= 0) {
          pickGoal(fish);
        }

        const speed =
          fish.cruise * (fish.mode === "dart" ? DART_SPEED : fish.mode === "idle" ? IDLE_SPEED : 1);
        const wantX = distance > 0 ? (toX / distance) * speed : 0;
        const wantY = distance > 0 ? (toY / distance) * speed * 0.7 : 0;
        const bend = STEER * (fish.mode === "dart" ? 3 : 1);
        fish.vx = approach(fish.vx, wantX, bend, dt);
        fish.vy = approach(fish.vy, wantY, bend, dt);

        const toZ = fish.goalZ - fish.z;
        fish.z += Math.sign(toZ) * Math.min(Math.abs(toZ), DEPTH_RATE * dt);
      }

      for (let i = 0; i < swimmers.length; i++) {
        for (let j = i + 1; j < swimmers.length; j++) {
          const a = swimmers[i]!;
          const b = swimmers[j]!;
          if (!a.shown || !b.shown || Math.abs(a.z - b.z) > SPACE_DEPTH) {
            continue;
          }
          const dx = b.x - a.x;
          const dy = b.y - a.y;
          const d = Math.hypot(dx, dy);
          if (d > 0 && d < PERSONAL_SPACE) {
            const push = ((PERSONAL_SPACE - d) / PERSONAL_SPACE) * SPACE_PUSH * dt;
            a.vx -= (dx / d) * push;
            a.vy -= (dy / d) * push;
            b.vx += (dx / d) * push;
            b.vy += (dy / d) * push;
          }
        }
      }

      for (const fish of swimmers) {
        const onScreen = mix(DEPTH_SPEED, fish.z);
        fish.x = Math.min(maxX(), Math.max(EDGE, fish.x + fish.vx * onScreen * dt));
        fish.y = Math.min(maxY(), Math.max(EDGE, fish.y + fish.vy * onScreen * dt));

        if (fish.vx > TURN_THRESHOLD) {
          fish.facing = 1;
        } else if (fish.vx < -TURN_THRESHOLD) {
          fish.facing = -1;
        }
        fish.turn = approach(fish.turn, fish.facing, TURN_RATE, dt);

        const heading = (Math.atan2(fish.vy, Math.max(Math.abs(fish.vx), 0.5)) * 180) / Math.PI;
        const wantTilt = Math.max(-MAX_TILT, Math.min(MAX_TILT, heading));
        fish.tilt = approach(fish.tilt, wantTilt, TILT_RATE, dt);

        const speed = Math.hypot(fish.vx, fish.vy);
        fish.tailPhase += (TAIL_IDLE + speed * TAIL_PER_SPEED) * dt;
        fish.bobPhase += dt * 1.6;
      }
    }

    function draw() {
      swimmers.forEach((fish, i) => {
        const el = fishEls[i];
        const { turn, frame } = parts[i] ?? {};
        if (!el) {
          return;
        }
        const bob = Math.sin(fish.bobPhase) * 0.8;
        const scale = mix(DEPTH_SCALE, fish.z);
        el.style.transform =
          `translate(calc(var(--px) * ${fish.x.toFixed(2)}), calc(var(--px) * ${(fish.y + bob).toFixed(2)}))` +
          ` scale(${scale.toFixed(3)})`;
        el.style.opacity = mix(DEPTH_OPACITY, fish.z).toFixed(3);
        // The tank's front layer is at 10: only the nearest fish pass it.
        el.style.zIndex = String(1 + Math.round(fish.z * 10));

        if (turn) {
          // Rotate before the flip, so a left-facing fish dips its own nose.
          const sway = Math.sin(fish.tailPhase * (Math.PI / 2)) * SWAY;
          turn.style.transform = `scaleX(${fish.turn.toFixed(3)}) rotate(${(fish.tilt + sway).toFixed(2)}deg)`;
        }
        if (frame) {
          frame.dataset.frame = String(
            TAIL_FRAMES[Math.floor(fish.tailPhase) % TAIL_FRAMES.length],
          );
        }
      });
    }

    function reset() {
      delete water!.dataset.live;
      fishEls.forEach((el, i) => {
        el.style.removeProperty("transform");
        el.style.removeProperty("opacity");
        el.style.removeProperty("z-index");
        parts[i]?.turn?.style.removeProperty("transform");
      });
    }

    let rafId = 0;
    let last = 0;
    let visible = true;

    function frame(now: number) {
      const dt = last ? Math.min(MAX_DT, (now - last) / 1000) : 0;
      last = now;
      step(dt);
      draw();
      rafId = requestAnimationFrame(frame);
    }

    function start() {
      if (rafId || reducedMotion.matches || !visible || swimmers.length === 0) {
        return;
      }
      water!.dataset.live = "";
      last = 0;
      draw();
      rafId = requestAnimationFrame(frame);
    }

    function stop() {
      cancelAnimationFrame(rafId);
      rafId = 0;
    }

    function onMotionChange() {
      if (reducedMotion.matches) {
        stop();
        reset();
      } else {
        start();
      }
    }

    const resize = new ResizeObserver(measure);
    resize.observe(water);
    const onScreen = new IntersectionObserver(([entry]) => {
      visible = entry?.isIntersecting ?? true;
      if (visible) {
        start();
      } else {
        stop();
      }
    });
    onScreen.observe(water);
    reducedMotion.addEventListener("change", onMotionChange);

    start();

    return () => {
      stop();
      resize.disconnect();
      onScreen.disconnect();
      reducedMotion.removeEventListener("change", onMotionChange);
      reset();
    };
  }, [waterRef, startsKey]);
}
