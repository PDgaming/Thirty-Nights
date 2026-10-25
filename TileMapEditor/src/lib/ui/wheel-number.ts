/**
 * Mouse-wheel scrubbing for `<input type="number">` fields.
 *
 * Wheel deltas arrive in pixels, lines, or pages depending on the browser and
 * device. Trackpads also emit many small deltas per gesture, so deltas are
 * accumulated until one field step is worth committing. Resetting the
 * accumulator after a step keeps a mouse wheel at one step per notch.
 */

export const WHEEL_STEP_PIXELS = 40;

const LINE_HEIGHT = 16;
const PAGE_HEIGHT = 400;

export function wheelDeltaPixels(deltaY: number, deltaMode: number): number {
	if (deltaMode === 1) return deltaY * LINE_HEIGHT;
	if (deltaMode === 2) return deltaY * PAGE_HEIGHT;
	return deltaY;
}

export type WheelStep = {
	/** -1 to step down, 1 to step up, 0 while the gesture hasn't added up to a step. */
	direction: -1 | 0 | 1;
	accumulated: number;
};

export function resolveWheelStep(
	accumulated: number,
	deltaY: number,
	deltaMode: number
): WheelStep {
	const pixels = wheelDeltaPixels(deltaY, deltaMode);
	if (pixels === 0) return { direction: 0, accumulated };

	// A direction change starts a fresh gesture instead of cancelling out old
	// momentum from the opposite direction.
	const total = Math.sign(pixels) === Math.sign(accumulated) ? accumulated + pixels : pixels;

	if (Math.abs(total) < WHEEL_STEP_PIXELS) {
		return { direction: 0, accumulated: total };
	}

	return { direction: total > 0 ? -1 : 1, accumulated: 0 };
}

function dispatchInputEvents(input: HTMLInputElement) {
	input.dispatchEvent(new Event('input', { bubbles: true }));
	input.dispatchEvent(new Event('change', { bubbles: true }));
}

/**
 * Svelte action: hovering a number field and scrolling changes its value.
 * The field's own `min`, `max`, and `step` attributes are honoured.
 */
export function wheelNumber(input: HTMLInputElement) {
	let accumulated = 0;

	function onWheel(event: WheelEvent) {
		if (input.disabled || input.readOnly) return;

		const { direction, accumulated: nextAccumulated } = resolveWheelStep(
			accumulated,
			event.deltaY,
			event.deltaMode
		);
		accumulated = nextAccumulated;

		// Claim the gesture even while it is still accumulating so trackpad
		// scrolling does not move the panel before a step fires.
		event.preventDefault();

		if (direction === 0) return;

		const before = input.value;
		if (direction > 0) input.stepUp();
		else input.stepDown();
		if (input.value === before) return;

		dispatchInputEvents(input);
	}

	input.addEventListener('wheel', onWheel, { passive: false });

	return {
		destroy() {
			input.removeEventListener('wheel', onWheel);
		}
	};
}
