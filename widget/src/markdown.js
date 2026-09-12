// The first ~8 words of the note become the report heading. Empty input yields
// an empty string so callers can supply their own fallback.
export function noteTitle(note) {
  const words = String(note).trim().split(/\s+/).filter(Boolean);
  return words.slice(0, 8).join(" ");
}

function pad(n, width = 2) {
  return String(n).padStart(width, "0");
}

// Filesystem-safe (no colons/slashes) timestamp for export filenames. Includes
// milliseconds so repeated exports within the same second never overwrite.
export function formatExportStamp(date) {
  const d = new Date(date);
  return (
    d.getUTCFullYear() +
    pad(d.getUTCMonth() + 1) +
    pad(d.getUTCDate()) +
    "-" +
    pad(d.getUTCHours()) +
    pad(d.getUTCMinutes()) +
    pad(d.getUTCSeconds()) +
    "-" +
    pad(d.getUTCMilliseconds(), 3)
  );
}

export function exportFilename(date) {
  return `bugs-${formatExportStamp(date)}.md`;
}

function formatPosition(p) {
  const x = Math.round(p.xPercent * 100);
  const y = Math.round(p.yPercent * 100);
  return `- Position: ${x}% from left, ${y}% from top (viewport ${p.viewportWidth}x${p.viewportHeight})`;
}

export function toMarkdown(flags, exportedAt = new Date()) {
  const lines = [
    "# Squawk Bug Report",
    `Exported: ${exportedAt.toISOString()}`,
    "",
  ];

  if (!flags.length) {
    lines.push("_No bugs flagged._");
    return lines.join("\n") + "\n";
  }

  for (const f of flags) {
    const title = noteTitle(f.note) || "(no note)";
    lines.push("", `## ${f.id} — ${title}`);
    lines.push(`- Page: ${f.url}`);
    lines.push(`- Time: ${f.timestamp}`);
    if (f.position) lines.push(formatPosition(f.position));
    lines.push("");
    if (f.note) lines.push(f.note);
  }

  return lines.join("\n") + "\n";
}