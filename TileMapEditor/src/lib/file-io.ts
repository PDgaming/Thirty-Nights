import type { TilesetSlice } from './tileset/slicing';

export interface LoadedImage {
	image: HTMLImageElement;
	objectUrl: string;
	width: number;
	height: number;
}

export async function decodeImageFile(file: File): Promise<LoadedImage> {
	if (!file.type.startsWith('image/')) {
		throw new Error(`"${file.name}" is not an image file`);
	}

	const objectUrl = URL.createObjectURL(file);
	const image = new Image();
	image.src = objectUrl;

	try {
		await image.decode();
	} catch {
		URL.revokeObjectURL(objectUrl);
		throw new Error(`Could not decode "${file.name}". Try a PNG or JPEG file.`);
	}

	if (image.naturalWidth === 0 || image.naturalHeight === 0) {
		URL.revokeObjectURL(objectUrl);
		throw new Error(`"${file.name}" has no readable image dimensions`);
	}

	return {
		image,
		objectUrl,
		width: image.naturalWidth,
		height: image.naturalHeight
	};
}

function toHexChannel(value: number): string {
	return value.toString(16).padStart(2, '0');
}

/**
 * Sample the middle pixel of every slice in one pass. These colours become the
 * game tile definitions, because the game schema renders colour rather than the
 * source image. The image itself stays an editor-side reference.
 */
export function extractSliceColors(
	image: HTMLImageElement,
	slices: TilesetSlice[]
): Record<number, string> {
	const colors: Record<number, string> = {};
	const width = image.naturalWidth;
	const height = image.naturalHeight;

	if (slices.length === 0 || width === 0 || height === 0) return colors;

	const canvas = document.createElement('canvas');
	canvas.width = width;
	canvas.height = height;
	const context = canvas.getContext('2d', { willReadFrequently: true });
	if (!context) return colors;

	try {
		context.drawImage(image, 0, 0);
		const pixels = context.getImageData(0, 0, width, height).data;
		for (const slice of slices) {
			const sampleX = Math.min(width - 1, Math.floor(slice.sx + slice.sw / 2));
			const sampleY = Math.min(height - 1, Math.floor(slice.sy + slice.sh / 2));
			const offset = (sampleY * width + sampleX) * 4;
			colors[slice.index] =
				`#${toHexChannel(pixels[offset])}${toHexChannel(pixels[offset + 1])}${toHexChannel(pixels[offset + 2])}`;
		}
	} catch {
		// A tainted or oversized canvas is not fatal: tiles still get a valid
		// default colour and can be edited in the inspector.
	}

	return colors;
}

export function downloadJson(filename: string, json: string): void {
	const blob = new Blob([json], { type: 'application/json' });
	const url = URL.createObjectURL(blob);
	const link = document.createElement('a');
	link.href = url;
	link.download = filename;
	document.body.appendChild(link);
	link.click();
	link.remove();
	URL.revokeObjectURL(url);
}
