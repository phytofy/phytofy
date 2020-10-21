// Copyright (c) 2020 OSRAM; Licensed under the MIT license.

export const ELEVATION_DEFAULT = 4;

export const SPACING_DEFAULT = 2;

export const FORMATION_2_BY_2 = 0;
export const FORMATION_1 = 1;
export const FORMATION_2 = 2;
export const FORMATION_3 = 3;
export const FORMATION_4 = 4;

export const FORMATIONS = [
  { text: "2x2", value: FORMATION_2_BY_2, countX: 2, countY: 2, orientation: 0 },
  { text: "1", value: FORMATION_1, countX: 1, countY: 1, orientation: 0 },
  { text: "2", value: FORMATION_2, countX: 2, countY: 1, orientation: 1 },
  { text: "3", value: FORMATION_3, countX: 3, countY: 1, orientation: 1 },
  { text: "4", value: FORMATION_4, countX: 4, countY: 1, orientation: 1 },
];

export const FORMATION_DEFAULT = FORMATION_3;

export const CHANNEL_UV_A = 0;
export const CHANNEL_BLUE = 1;
export const CHANNEL_GREEN = 2;
export const CHANNEL_HYPER_RED = 3;
export const CHANNEL_FAR_RED = 4;
export const CHANNEL_WHITE = 5;

export const CHANNELS = [
  { text: "UV-A", value: CHANNEL_UV_A },
  { text: "Blue", value: CHANNEL_BLUE },
  { text: "Green", value: CHANNEL_GREEN },
  { text: "Hyper Red", value: CHANNEL_HYPER_RED },
  { text: "Far Red", value: CHANNEL_FAR_RED },
  { text: "White", value: CHANNEL_WHITE },
];

export const CHANNEL_DEFAULT = CHANNEL_WHITE;

export const SCALES = ["Rainbow", "Electric", "Greys"];

export const SCALE_DEFAULT = SCALES[1];

export const LUMINAIRES_VISIBLE = true;
export const LUMINAIRES_HIDDEN = false;

export const LUMINAIRES = [
  { text: "visible", value: LUMINAIRES_VISIBLE },
  { text: "hidden", value: LUMINAIRES_HIDDEN },
];

export const LUMINAIRES_DEFAULT = LUMINAIRES_VISIBLE;
