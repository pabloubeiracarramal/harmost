const relativeTimeFormatter = new Intl.RelativeTimeFormat('en', { numeric: 'auto' });

export function formatRelativeTime(isoDate: string): string {
  const diffMs = new Date(isoDate).getTime() - Date.now();
  const diffSeconds = Math.round(diffMs / 1000);
  const units: [Intl.RelativeTimeFormatUnit, number][] = [
    ['year', 60 * 60 * 24 * 365],
    ['month', 60 * 60 * 24 * 30],
    ['day', 60 * 60 * 24],
    ['hour', 60 * 60],
    ['minute', 60],
  ];
  for (const [unit, secondsInUnit] of units) {
    if (Math.abs(diffSeconds) >= secondsInUnit) {
      return relativeTimeFormatter.format(Math.round(diffSeconds / secondsInUnit), unit);
    }
  }
  return relativeTimeFormatter.format(diffSeconds, 'second');
}
