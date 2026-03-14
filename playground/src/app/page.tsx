"use client";

import { useState } from "react";
import SchemaEditor from "@/components/SchemaEditor";
import ResultsTable from "@/components/ResultsTable";
import ImpactPanel from "@/components/ImpactPanel";
import { SchemaType, DiffResult } from "@/lib/types";
import { SAMPLES } from "@/lib/samples";

const SCHEMA_TYPES: { id: SchemaType; label: string }[] = [
  { id: "openapi",  label: "OpenAPI" },
  { id: "graphql",  label: "GraphQL" },
  { id: "grpc",     label: "gRPC / Protobuf" },
];

type ResultTab = "diff" | "impact";

const API_URL = "";

export default function PlaygroundPage() {
  const [schemaType, setSchemaType] = useState<SchemaType>("openapi");
  const [base, setBase] = useState(SAMPLES.openapi.base);
  const [head, setHead] = useState(SAMPLES.openapi.head);
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<DiffResult | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<ResultTab>("diff");

  function switchType(type: SchemaType) {
    setSchemaType(type);
    setBase(SAMPLES[type].base);
    setHead(SAMPLES[type].head);
    setResult(null);
    setError(null);
    setActiveTab("diff");
  }

  async function handleCompare() {
    setLoading(true);
    setResult(null);
    setError(null);
    setActiveTab("diff");

    try {
      const res = await fetch(`${API_URL}/api/compare`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ schema_type: schemaType, base_content: base, head_content: head }),
      });

      const data = await res.json();

      if (!res.ok) {
        setError(data.error ?? "Unknown error");
        return;
      }

      setResult(data as DiffResult);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Network error");
    } finally {
      setLoading(false);
    }
  }

  const hasBreaking = (result?.summary.breaking ?? 0) > 0;

  return (
    <div className="min-h-screen bg-white text-gray-900">
      {/* Header */}
      <header className="border-b border-gray-200 bg-white shadow-sm">
        <div className="max-w-[1400px] mx-auto px-6 py-4 flex items-center justify-between flex-wrap gap-4">
          <div className="flex items-center gap-3">
            <span className="text-xl font-bold">
              <span className="text-indigo-600">drift-guard</span> playground
            </span>
            <span className="text-[11px] font-medium bg-gray-100 text-gray-500 px-2 py-0.5 rounded-full border border-gray-200">
              beta
            </span>
          </div>

          <a
            href="https://pgomes13.github.io/api-drift-engine/generating-specs"
            target="_blank"
            rel="noopener noreferrer"
            className="text-sm font-medium text-indigo-600 hover:text-indigo-800 border border-indigo-200 hover:border-indigo-400 px-3 py-1.5 rounded-lg transition-colors"
          >
            Generate specs ↗
          </a>

          <div className="flex gap-1 bg-gray-100 rounded-lg p-1">
            {SCHEMA_TYPES.map((t) => (
              <button
                key={t.id}
                onClick={() => switchType(t.id)}
                className={`px-4 py-1.5 rounded-md text-sm font-medium transition-all ${
                  schemaType === t.id
                    ? "bg-indigo-600 text-white shadow-sm"
                    : "text-gray-500 hover:text-gray-700 hover:bg-white"
                }`}
              >
                {t.label}
              </button>
            ))}
          </div>
        </div>
      </header>

      <main className="max-w-[1400px] mx-auto px-6 py-6 space-y-5">
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-5">
          <SchemaEditor
            label="Base Schema"
            subtitle="— current / stable version"
            value={base}
            onChange={setBase}
            schemaType={schemaType}
          />
          <SchemaEditor
            label="Head Schema"
            subtitle="— new / proposed version"
            value={head}
            onChange={setHead}
            schemaType={schemaType}
          />
        </div>

        <div className="flex justify-center">
          <button
            onClick={handleCompare}
            disabled={loading}
            className="px-10 py-2.5 bg-indigo-600 hover:bg-indigo-700 disabled:opacity-60 disabled:cursor-not-allowed text-white font-semibold rounded-lg transition-colors text-base shadow-sm"
          >
            {loading ? (
              <span className="flex items-center gap-2">
                <svg className="animate-spin h-4 w-4" viewBox="0 0 24 24" fill="none">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8H4z" />
                </svg>
                Comparing…
              </span>
            ) : "Compare"}
          </button>
        </div>

        {error && (
          <div className="rounded-lg border border-red-300 bg-red-50 px-4 py-3 text-red-700 text-sm font-mono whitespace-pre-wrap">
            {error}
          </div>
        )}

        {result && (
          <div className="space-y-4">
            {/* Tab bar — Impact tab only shown when there are breaking changes */}
            <div className="flex gap-1 border-b border-gray-200">
              <button
                onClick={() => setActiveTab("diff")}
                className={`px-4 py-2 text-sm font-medium border-b-2 -mb-px transition-colors ${
                  activeTab === "diff"
                    ? "border-indigo-600 text-indigo-600"
                    : "border-transparent text-gray-500 hover:text-gray-700"
                }`}
              >
                Diff Results
              </button>
              {hasBreaking && (
                <button
                  onClick={() => setActiveTab("impact")}
                  className={`px-4 py-2 text-sm font-medium border-b-2 -mb-px transition-colors flex items-center gap-1.5 ${
                    activeTab === "impact"
                      ? "border-red-600 text-red-600"
                      : "border-transparent text-gray-500 hover:text-gray-700"
                  }`}
                >
                  Impact Analysis
                  <span className="px-1.5 py-0.5 rounded-full text-[10px] font-bold bg-red-100 text-red-700">
                    {result.summary.breaking}
                  </span>
                </button>
              )}
            </div>

            {activeTab === "diff" && <ResultsTable result={result} />}
            {activeTab === "impact" && (
              <ImpactPanel diff={result} apiUrl={API_URL} />
            )}
          </div>
        )}
      </main>
    </div>
  );
}
