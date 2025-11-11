import chroma from "chroma-js";

const red = {
  50: chroma.oklch(0.971, 0.013, 17.38),
  100: chroma.oklch(0.936, 0.032, 17.717),
  200: chroma.oklch(0.885, 0.062, 18.334),
  300: chroma.oklch(0.808, 0.114, 19.571),
  400: chroma.oklch(0.704, 0.191, 22.216),
  500: chroma.oklch(0.637, 0.237, 25.331),
  600: chroma.oklch(0.577, 0.245, 27.325),
  700: chroma.oklch(0.505, 0.213, 27.518),
  800: chroma.oklch(0.444, 0.177, 26.899),
  900: chroma.oklch(0.396, 0.141, 25.723),
  950: chroma.oklch(0.258, 0.092, 26.042),
};

const orange = {
  50: chroma.oklch(0.98, 0.016, 73.684),
  100: chroma.oklch(0.954, 0.038, 75.164),
  200: chroma.oklch(0.901, 0.076, 70.697),
  300: chroma.oklch(0.837, 0.128, 66.29),
  400: chroma.oklch(0.75, 0.183, 55.934),
  500: chroma.oklch(0.705, 0.213, 47.604),
  600: chroma.oklch(0.646, 0.222, 41.116),
  700: chroma.oklch(0.553, 0.195, 38.402),
  800: chroma.oklch(0.47, 0.157, 37.304),
  900: chroma.oklch(0.408, 0.123, 38.172),
  950: chroma.oklch(0.266, 0.079, 36.259),
};

const amber = {
  50: chroma.oklch(0.987, 0.022, 95.277),
  100: chroma.oklch(0.962, 0.059, 95.617),
  200: chroma.oklch(0.924, 0.12, 95.746),
  300: chroma.oklch(0.879, 0.169, 91.605),
  400: chroma.oklch(0.828, 0.189, 84.429),
  500: chroma.oklch(0.769, 0.188, 70.08),
  600: chroma.oklch(0.666, 0.179, 58.318),
  700: chroma.oklch(0.555, 0.163, 48.998),
  800: chroma.oklch(0.473, 0.137, 46.201),
  900: chroma.oklch(0.414, 0.112, 45.904),
  950: chroma.oklch(0.279, 0.077, 45.635),
};

const yellow = {
  50: chroma.oklch(0.987, 0.026, 102.212),
  100: chroma.oklch(0.973, 0.071, 103.193),
  200: chroma.oklch(0.945, 0.129, 101.54),
  300: chroma.oklch(0.905, 0.182, 98.111),
  400: chroma.oklch(0.852, 0.199, 91.936),
  500: chroma.oklch(0.795, 0.184, 86.047),
  600: chroma.oklch(0.681, 0.162, 75.834),
  700: chroma.oklch(0.554, 0.135, 66.442),
  800: chroma.oklch(0.476, 0.114, 61.907),
  900: chroma.oklch(0.421, 0.095, 57.708),
  950: chroma.oklch(0.286, 0.066, 53.813),
};

const lime = {
  50: chroma.oklch(0.986, 0.031, 120.757),
  100: chroma.oklch(0.967, 0.067, 122.328),
  200: chroma.oklch(0.938, 0.127, 124.321),
  300: chroma.oklch(0.897, 0.196, 126.665),
  400: chroma.oklch(0.841, 0.238, 128.85),
  500: chroma.oklch(0.768, 0.233, 130.85),
  600: chroma.oklch(0.648, 0.2, 131.684),
  700: chroma.oklch(0.532, 0.157, 131.589),
  800: chroma.oklch(0.453, 0.124, 130.933),
  900: chroma.oklch(0.405, 0.101, 131.063),
  950: chroma.oklch(0.274, 0.072, 132.109),
};

const green = {
  50: chroma.oklch(0.982, 0.018, 155.826),
  100: chroma.oklch(0.962, 0.044, 156.743),
  200: chroma.oklch(0.925, 0.084, 155.995),
  300: chroma.oklch(0.871, 0.15, 154.449),
  400: chroma.oklch(0.792, 0.209, 151.711),
  500: chroma.oklch(0.723, 0.219, 149.579),
  600: chroma.oklch(0.627, 0.194, 149.214),
  700: chroma.oklch(0.527, 0.154, 150.069),
  800: chroma.oklch(0.448, 0.119, 151.328),
  900: chroma.oklch(0.393, 0.095, 152.535),
  950: chroma.oklch(0.266, 0.065, 152.934),
};

const emerald = {
  50: chroma.oklch(0.979, 0.021, 166.113),
  100: chroma.oklch(0.95, 0.052, 163.051),
  200: chroma.oklch(0.905, 0.093, 164.15),
  300: chroma.oklch(0.845, 0.143, 164.978),
  400: chroma.oklch(0.765, 0.177, 163.223),
  500: chroma.oklch(0.696, 0.17, 162.48),
  600: chroma.oklch(0.596, 0.145, 163.225),
  700: chroma.oklch(0.508, 0.118, 165.612),
  800: chroma.oklch(0.432, 0.095, 166.913),
  900: chroma.oklch(0.378, 0.077, 168.94),
  950: chroma.oklch(0.262, 0.051, 172.552),
};

const teal = {
  50: chroma.oklch(0.984, 0.014, 180.72),
  100: chroma.oklch(0.953, 0.051, 180.801),
  200: chroma.oklch(0.91, 0.096, 180.426),
  300: chroma.oklch(0.855, 0.138, 181.071),
  400: chroma.oklch(0.777, 0.152, 181.912),
  500: chroma.oklch(0.704, 0.14, 182.503),
  600: chroma.oklch(0.6, 0.118, 184.704),
  700: chroma.oklch(0.511, 0.096, 186.391),
  800: chroma.oklch(0.437, 0.078, 188.216),
  900: chroma.oklch(0.386, 0.063, 188.416),
  950: chroma.oklch(0.277, 0.046, 192.524),
};

const cyan = {
  50: chroma.oklch(0.984, 0.019, 200.873),
  100: chroma.oklch(0.956, 0.045, 203.388),
  200: chroma.oklch(0.917, 0.08, 205.041),
  300: chroma.oklch(0.865, 0.127, 207.078),
  400: chroma.oklch(0.789, 0.154, 211.53),
  500: chroma.oklch(0.715, 0.143, 215.221),
  600: chroma.oklch(0.609, 0.126, 221.723),
  700: chroma.oklch(0.52, 0.105, 223.128),
  800: chroma.oklch(0.45, 0.085, 224.283),
  900: chroma.oklch(0.398, 0.07, 227.392),
  950: chroma.oklch(0.302, 0.056, 229.695),
};

const sky = {
  50: chroma.oklch(0.977, 0.013, 236.62),
  100: chroma.oklch(0.951, 0.026, 236.824),
  200: chroma.oklch(0.901, 0.058, 230.902),
  300: chroma.oklch(0.828, 0.111, 230.318),
  400: chroma.oklch(0.746, 0.16, 232.661),
  500: chroma.oklch(0.685, 0.169, 237.323),
  600: chroma.oklch(0.588, 0.158, 241.966),
  700: chroma.oklch(0.5, 0.134, 242.749),
  800: chroma.oklch(0.443, 0.11, 240.79),
  900: chroma.oklch(0.391, 0.09, 240.876),
  950: chroma.oklch(0.293, 0.066, 243.157),
};

const blue = {
  50: chroma.oklch(0.97, 0.014, 254.604),
  100: chroma.oklch(0.932, 0.032, 255.585),
  200: chroma.oklch(0.882, 0.059, 254.128),
  300: chroma.oklch(0.809, 0.105, 251.813),
  400: chroma.oklch(0.707, 0.165, 254.624),
  500: chroma.oklch(0.623, 0.214, 259.815),
  600: chroma.oklch(0.546, 0.245, 262.881),
  700: chroma.oklch(0.488, 0.243, 264.376),
  800: chroma.oklch(0.424, 0.199, 265.638),
  900: chroma.oklch(0.379, 0.146, 265.522),
  950: chroma.oklch(0.282, 0.091, 267.935),
};

const indigo = {
  50: chroma.oklch(0.962, 0.018, 272.314),
  100: chroma.oklch(0.93, 0.034, 272.788),
  200: chroma.oklch(0.87, 0.065, 274.039),
  300: chroma.oklch(0.785, 0.115, 274.713),
  400: chroma.oklch(0.673, 0.182, 276.935),
  500: chroma.oklch(0.585, 0.233, 277.117),
  600: chroma.oklch(0.511, 0.262, 276.966),
  700: chroma.oklch(0.457, 0.24, 277.023),
  800: chroma.oklch(0.398, 0.195, 277.366),
  900: chroma.oklch(0.359, 0.144, 278.697),
  950: chroma.oklch(0.257, 0.09, 281.288),
};

const violet = {
  50: chroma.oklch(0.969, 0.016, 293.756),
  100: chroma.oklch(0.943, 0.029, 294.588),
  200: chroma.oklch(0.894, 0.057, 293.283),
  300: chroma.oklch(0.811, 0.111, 293.571),
  400: chroma.oklch(0.702, 0.183, 293.541),
  500: chroma.oklch(0.606, 0.25, 292.717),
  600: chroma.oklch(0.541, 0.281, 293.009),
  700: chroma.oklch(0.491, 0.27, 292.581),
  800: chroma.oklch(0.432, 0.232, 292.759),
  900: chroma.oklch(0.38, 0.189, 293.745),
  950: chroma.oklch(0.283, 0.141, 291.089),
};

const purple = {
  50: chroma.oklch(0.977, 0.014, 308.299),
  100: chroma.oklch(0.946, 0.033, 307.174),
  200: chroma.oklch(0.902, 0.063, 306.703),
  300: chroma.oklch(0.827, 0.119, 306.383),
  400: chroma.oklch(0.714, 0.203, 305.504),
  500: chroma.oklch(0.627, 0.265, 303.9),
  600: chroma.oklch(0.558, 0.288, 302.321),
  700: chroma.oklch(0.496, 0.265, 301.924),
  800: chroma.oklch(0.438, 0.218, 303.724),
  900: chroma.oklch(0.381, 0.176, 304.987),
  950: chroma.oklch(0.291, 0.149, 302.717),
};

const fuchsia = {
  50: chroma.oklch(0.977, 0.017, 320.058),
  100: chroma.oklch(0.952, 0.037, 318.852),
  200: chroma.oklch(0.903, 0.076, 319.62),
  300: chroma.oklch(0.833, 0.145, 321.434),
  400: chroma.oklch(0.74, 0.238, 322.16),
  500: chroma.oklch(0.667, 0.295, 322.15),
  600: chroma.oklch(0.591, 0.293, 322.896),
  700: chroma.oklch(0.518, 0.253, 323.949),
  800: chroma.oklch(0.452, 0.211, 324.591),
  900: chroma.oklch(0.401, 0.17, 325.612),
  950: chroma.oklch(0.293, 0.136, 325.661),
};

const pink = {
  50: chroma.oklch(0.971, 0.014, 343.198),
  100: chroma.oklch(0.948, 0.028, 342.258),
  200: chroma.oklch(0.899, 0.061, 343.231),
  300: chroma.oklch(0.823, 0.12, 346.018),
  400: chroma.oklch(0.718, 0.202, 349.761),
  500: chroma.oklch(0.656, 0.241, 354.308),
  600: chroma.oklch(0.592, 0.249, 0.584),
  700: chroma.oklch(0.525, 0.223, 3.958),
  800: chroma.oklch(0.459, 0.187, 3.815),
  900: chroma.oklch(0.408, 0.153, 2.432),
  950: chroma.oklch(0.284, 0.109, 3.907),
};

const rose = {
  50: chroma.oklch(0.969, 0.015, 12.422),
  100: chroma.oklch(0.941, 0.03, 12.58),
  200: chroma.oklch(0.892, 0.058, 10.001),
  300: chroma.oklch(0.81, 0.117, 11.638),
  400: chroma.oklch(0.712, 0.194, 13.428),
  500: chroma.oklch(0.645, 0.246, 16.439),
  600: chroma.oklch(0.586, 0.253, 17.585),
  700: chroma.oklch(0.514, 0.222, 16.935),
  800: chroma.oklch(0.455, 0.188, 13.697),
  900: chroma.oklch(0.41, 0.159, 10.272),
  950: chroma.oklch(0.271, 0.105, 12.094),
};

const slate = {
  50: chroma.oklch(0.984, 0.003, 247.858),
  100: chroma.oklch(0.968, 0.007, 247.896),
  200: chroma.oklch(0.929, 0.013, 255.508),
  300: chroma.oklch(0.869, 0.022, 252.894),
  400: chroma.oklch(0.704, 0.04, 256.788),
  500: chroma.oklch(0.554, 0.046, 257.417),
  600: chroma.oklch(0.446, 0.043, 257.281),
  700: chroma.oklch(0.372, 0.044, 257.287),
  800: chroma.oklch(0.279, 0.041, 260.031),
  900: chroma.oklch(0.208, 0.042, 265.755),
  950: chroma.oklch(0.129, 0.042, 264.695),
};

const gray = {
  50: chroma.oklch(0.985, 0.002, 247.839),
  100: chroma.oklch(0.967, 0.003, 264.542),
  200: chroma.oklch(0.928, 0.006, 264.531),
  300: chroma.oklch(0.872, 0.01, 258.338),
  400: chroma.oklch(0.707, 0.022, 261.325),
  500: chroma.oklch(0.551, 0.027, 264.364),
  600: chroma.oklch(0.446, 0.03, 256.802),
  700: chroma.oklch(0.373, 0.034, 259.733),
  800: chroma.oklch(0.278, 0.033, 256.848),
  900: chroma.oklch(0.21, 0.034, 264.665),
  950: chroma.oklch(0.13, 0.028, 261.692),
};

const zinc = {
  50: chroma.oklch(0.985, 0, 0),
  100: chroma.oklch(0.967, 0.001, 286.375),
  200: chroma.oklch(0.92, 0.004, 286.32),
  300: chroma.oklch(0.871, 0.006, 286.286),
  400: chroma.oklch(0.705, 0.015, 286.067),
  500: chroma.oklch(0.552, 0.016, 285.938),
  600: chroma.oklch(0.442, 0.017, 285.786),
  700: chroma.oklch(0.37, 0.013, 285.805),
  800: chroma.oklch(0.274, 0.006, 286.033),
  900: chroma.oklch(0.21, 0.006, 285.885),
  950: chroma.oklch(0.141, 0.005, 285.823),
};

const neutral = {
  50: chroma.oklch(0.985, 0, 0),
  100: chroma.oklch(0.97, 0, 0),
  200: chroma.oklch(0.922, 0, 0),
  300: chroma.oklch(0.87, 0, 0),
  400: chroma.oklch(0.708, 0, 0),
  500: chroma.oklch(0.556, 0, 0),
  600: chroma.oklch(0.439, 0, 0),
  700: chroma.oklch(0.371, 0, 0),
  800: chroma.oklch(0.269, 0, 0),
  900: chroma.oklch(0.205, 0, 0),
  950: chroma.oklch(0.145, 0, 0),
};

const stone = {
  50: chroma.oklch(0.985, 0.001, 106.423),
  100: chroma.oklch(0.97, 0.001, 106.424),
  200: chroma.oklch(0.923, 0.003, 48.717),
  300: chroma.oklch(0.869, 0.005, 56.366),
  400: chroma.oklch(0.709, 0.01, 56.259),
  500: chroma.oklch(0.553, 0.013, 58.071),
  600: chroma.oklch(0.444, 0.011, 73.639),
  700: chroma.oklch(0.374, 0.01, 67.558),
  800: chroma.oklch(0.268, 0.007, 34.298),
  900: chroma.oklch(0.216, 0.006, 56.043),
  950: chroma.oklch(0.147, 0.004, 49.25),
};

export const primary = {
  50: chroma.oklch(0.96, 0.0206, 274.04),
  100: chroma.oklch(0.92, 0.0375, 274.1),
  200: chroma.oklch(0.86, 0.0707, 275.12),
  300: chroma.oklch(0.76, 0.1218, 275.41),
  400: chroma.oklch(0.65, 0.188451, 276.7938),
  500: chroma.oklch(0.56, 0.2439, 275.39),
  600: chroma.oklch(0.5, 0.2657, 275.04),
  700: chroma.oklch(0.42, 0.2319, 274.45),
  800: chroma.oklch(0.39, 0.2033, 275.31),
  900: chroma.oklch(0.35, 0.1574, 277.12),
  950: chroma.oklch(0.25, 0.1016, 280.81),
};

export const secondary = {
  50: chroma.oklch(0.98, 0.0134, 226.56),
  100: chroma.oklch(0.95, 0.0275, 232.66),
  200: chroma.oklch(0.91, 0.0603, 226.17),
  300: chroma.oklch(0.84, 0.1076, 224.29),
  400: chroma.oklch(0.78, 0.143, 227.54),
  500: chroma.oklch(0.74, 0.157556, 235.3407),
  600: chroma.oklch(0.62, 0.144066, 240.7481),
  700: chroma.oklch(0.53, 0.124077, 241.7448),
  800: chroma.oklch(0.47, 0.1045, 238.86),
  900: chroma.oklch(0.41, 0.088, 237.95),
  950: chroma.oklch(0.3, 0.0672, 241.09),
};

/**
 * Default Rill primary color palette as Color array
 */
export const defaultPrimaryPalette = Object.values(primary);

/**
 * Default Rill secondary color palette as Color array
 */
export const defaultSecondaryPalette = Object.values(secondary);

export const allColors = {
  red,
  orange,
  amber,
  yellow,
  lime,
  green,
  emerald,
  teal,
  cyan,
  sky,
  blue,
  indigo,
  violet,
  purple,
  fuchsia,
  pink,
  rose,
  gray,
  slate,
  zinc,
  neutral,
  stone,
  primary,
  secondary,
};
