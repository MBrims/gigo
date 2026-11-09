// Handy byte formatter. Rounds to 2 decimal places by default.
function FormatBytes(bytes: number, decimals?: number) {
  if (bytes == 0) return '0 Bytes';
  const k = 1024,
    dm = decimals || 2,
    sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'],
    i = Math.floor(Math.log(bytes) / Math.log(k));
  return (bytes / Math.pow(k, i)).toFixed(dm) + ' ' + sizes[i];
}

export {FormatBytes};