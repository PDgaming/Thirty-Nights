import { describe, expect, it } from 'vitest';
import { resolveWheelStep, wheelDeltaPixels, WHEEL_STEP_PIXELS } from './wheel-number';

describe('wheelDeltaPixels', () => {
	it('passes pixel deltas through', () => {
		expect(wheelDeltaPixels(12, 0)).toBe(12);
	});

	it('converts line deltas to pixels', () => {
		expect(wheelDeltaPixels(3, 1)).toBe(48);
	});

	it('converts page deltas to pixels', () => {
		expect(wheelDeltaPixels(1, 2)).toBe(400);
	});
});

describe('resolveWheelStep', () => {
	it('steps up when scrolling up', () => {
		expect(resolveWheelStep(0, -WHEEL_STEP_PIXELS, 0)).toEqual({
			direction: 1,
			accumulated: 0
		});
	});

	it('steps down when scrolling down', () => {
		expect(resolveWheelStep(0, WHEEL_STEP_PIXELS, 0)).toEqual({
			direction: -1,
			accumulated: 0
		});
	});

	it('accumulates small trackpad deltas until a step is reached', () => {
		const first = resolveWheelStep(0, 10, 0);
		expect(first.direction).toBe(0);

		const second = resolveWheelStep(first.accumulated, 10, 0);
		expect(second.direction).toBe(0);

		const third = resolveWheelStep(second.accumulated, 20, 0);
		expect(third).toEqual({ direction: -1, accumulated: 0 });
	});

	it('requires one notch per step rather than accelerating on large deltas', () => {
		const first = resolveWheelStep(0, 300, 0);
		expect(first).toEqual({ direction: -1, accumulated: 0 });

		const second = resolveWheelStep(first.accumulated, 300, 0);
		expect(second).toEqual({ direction: -1, accumulated: 0 });
	});

	it('treats a reversal as a new gesture', () => {
		const up = resolveWheelStep(0, 10, 0);
		const down = resolveWheelStep(up.accumulated, -WHEEL_STEP_PIXELS, 0);
		expect(down).toEqual({ direction: 1, accumulated: 0 });
	});

	it('ignores zero deltas', () => {
		expect(resolveWheelStep(15, 0, 0)).toEqual({ direction: 0, accumulated: 15 });
	});
});
