import { Change, DiffResult, Severity } from "@/lib/types";

interface Props {
  result: DiffResult;
}

const SEV_BADGE: Record<Severity, string> = {
  breaking:      "bg-red-950 text-red-400 border border-red-700",
  "non-breaking": "bg-amber-950 text-amber-400 border border-amber-700",
  info:          "bg-green-950 text-green-400 border border-green-700",
};

const ROW_BORDER: Record<Severity, string> = {
  breaking:      "border-l-2 border-l-red-500",
  "non-breaking": "border-l-2 border-l-amber-500",
  info:          "border-l-2 border-l-green-500",
};

function SeverityBadge({ severity }: { severity: Severity }) {
  return (
    <span className={`inline-flex items-center px-2 py-0.5 rounded text-[11px] font-bold whitespace-nowrap ${SEV_BADGE[severity]}`}>
      {severity.toUpperCase()}
    </span>
  );
}

function ChangeRow({ change }: { change: Change }) {
  return (
    <tr className={`border-b border-slate-800 hover:bg-slate-800/40 ${ROW_BORDER[change.severity]}`}>
      <td className="px-4 py-3 align-top">
        <SeverityBadge severity={change.severity} />
      </td>
      <td className="px-4 py-3 align-top font-mono text-xs text-slate-400 whitespace-nowrap">
        {change.type}
      </td>
      <td className="px-4 py-3 align-top font-mono text-sm font-medium whitespace-nowrap">
        {change.method && (
          <span className="text-xs font-bold text-slate-500 mr-1">{change.method}</span>
        )}
        {change.path}
      </td>
      <td className="px-4 py-3 align-top text-sm text-slate-300 leading-relaxed">
        {change.description}
        {(change.before || change.after) && (
          <div className="mt-1 font-mono text-xs text-slate-500">
            {change.before && <s className="mr-2">{change.before}</s>}
            {change.after && <span>→ {change.after}</span>}
          </div>
        )}
      </td>
    </tr>
  );
}

export default function ResultsTable({ result }: Props) {
  const { summary, changes } = result;

  return (
    <div className="space-y-4">
      {/* Summary pills */}
      <div className="flex items-center gap-3 flex-wrap">
        <span className="text-sm font-semibold text-slate-400">Results</span>
        <span className="px-3 py-1 rounded-full text-xs font-semibold bg-slate-700 text-slate-200">
          {summary.total} total
        </span>
        {summary.breaking > 0 && (
          <span className="px-3 py-1 rounded-full text-xs font-semibold bg-red-950 text-red-400 border border-red-700">
            {summary.breaking} breaking
          </span>
        )}
        {summary.non_breaking > 0 && (
          <span className="px-3 py-1 rounded-full text-xs font-semibold bg-amber-950 text-amber-400 border border-amber-700">
            {summary.non_breaking} non-breaking
          </span>
        )}
        {summary.info > 0 && (
          <span className="px-3 py-1 rounded-full text-xs font-semibold bg-green-950 text-green-400 border border-green-700">
            {summary.info} info
          </span>
        )}
      </div>

      {/* Changes */}
      {changes.length === 0 ? (
        <div className="text-center py-12 text-green-400 font-medium">
          ✓ &nbsp;No changes detected
        </div>
      ) : (
        <div className="border border-slate-700 rounded-lg overflow-hidden overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="bg-slate-800 border-b border-slate-700">
                <th className="px-4 py-2.5 text-left text-[11px] font-semibold uppercase tracking-wider text-slate-400">Severity</th>
                <th className="px-4 py-2.5 text-left text-[11px] font-semibold uppercase tracking-wider text-slate-400">Type</th>
                <th className="px-4 py-2.5 text-left text-[11px] font-semibold uppercase tracking-wider text-slate-400">Path</th>
                <th className="px-4 py-2.5 text-left text-[11px] font-semibold uppercase tracking-wider text-slate-400">Description</th>
              </tr>
            </thead>
            <tbody>
              {changes.map((c, i) => <ChangeRow key={i} change={c} />)}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
