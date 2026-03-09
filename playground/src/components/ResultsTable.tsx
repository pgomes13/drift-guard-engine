import { Change, DiffResult, Severity } from "@/lib/types";

interface Props {
  result: DiffResult;
}

const SEV_BADGE: Record<Severity, string> = {
  breaking:       "bg-red-50 text-red-700 border border-red-300",
  "non-breaking": "bg-amber-50 text-amber-700 border border-amber-300",
  info:           "bg-green-50 text-green-700 border border-green-300",
};

const ROW_BORDER: Record<Severity, string> = {
  breaking:       "border-l-2 border-l-red-400",
  "non-breaking": "border-l-2 border-l-amber-400",
  info:           "border-l-2 border-l-green-400",
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
    <tr className={`border-b border-gray-100 hover:bg-gray-50 ${ROW_BORDER[change.severity]}`}>
      <td className="px-4 py-3 align-top">
        <SeverityBadge severity={change.severity} />
      </td>
      <td className="px-4 py-3 align-top font-mono text-xs text-gray-500 whitespace-nowrap">
        {change.type}
      </td>
      <td className="px-4 py-3 align-top font-mono text-sm font-medium whitespace-nowrap text-gray-800">
        {change.method && (
          <span className="text-xs font-bold text-gray-400 mr-1">{change.method}</span>
        )}
        {change.path}
      </td>
      <td className="px-4 py-3 align-top text-sm text-gray-700 leading-relaxed">
        {change.description}
        {(change.before || change.after) && (
          <div className="mt-1 font-mono text-xs text-gray-400">
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
        <span className="text-sm font-semibold text-gray-500">Results</span>
        <span className="px-3 py-1 rounded-full text-xs font-semibold bg-gray-100 text-gray-700 border border-gray-200">
          {summary.total} total
        </span>
        {summary.breaking > 0 && (
          <span className="px-3 py-1 rounded-full text-xs font-semibold bg-red-50 text-red-700 border border-red-300">
            {summary.breaking} breaking
          </span>
        )}
        {summary.non_breaking > 0 && (
          <span className="px-3 py-1 rounded-full text-xs font-semibold bg-amber-50 text-amber-700 border border-amber-300">
            {summary.non_breaking} non-breaking
          </span>
        )}
        {summary.info > 0 && (
          <span className="px-3 py-1 rounded-full text-xs font-semibold bg-green-50 text-green-700 border border-green-300">
            {summary.info} info
          </span>
        )}
      </div>

      {/* Changes */}
      {changes.length === 0 ? (
        <div className="text-center py-12 text-green-600 font-medium">
          ✓ &nbsp;No changes detected
        </div>
      ) : (
        <div className="border border-gray-200 rounded-lg overflow-hidden overflow-x-auto shadow-sm">
          <table className="w-full text-sm">
            <thead>
              <tr className="bg-gray-50 border-b border-gray-200">
                <th className="px-4 py-2.5 text-left text-[11px] font-semibold uppercase tracking-wider text-gray-500">Severity</th>
                <th className="px-4 py-2.5 text-left text-[11px] font-semibold uppercase tracking-wider text-gray-500">Type</th>
                <th className="px-4 py-2.5 text-left text-[11px] font-semibold uppercase tracking-wider text-gray-500">Path</th>
                <th className="px-4 py-2.5 text-left text-[11px] font-semibold uppercase tracking-wider text-gray-500">Description</th>
              </tr>
            </thead>
            <tbody className="bg-white">
              {changes.map((c, i) => <ChangeRow key={i} change={c} />)}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
