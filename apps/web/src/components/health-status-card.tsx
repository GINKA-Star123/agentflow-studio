import { Activity, AlertTriangle, CheckCircle2, XCircle } from "lucide-react";

import type { ServiceHealth } from "@/lib/health";

type HealthStatusCardProps = {
  service: ServiceHealth;
};

export function HealthStatusCard({ service }: HealthStatusCardProps) {
  const stateView = getStateView(service.state);

  return (
    <section className="rounded-lg border border-slate-200 bg-white p-5 shadow-sm">
      <div className="flex items-start justify-between gap-4">
        <div className="min-w-0">
          <div className="flex items-center gap-2">
            <Activity className="h-4 w-4 text-slate-500" aria-hidden="true" />
            <h2 className="truncate text-base font-semibold text-slate-950">
              {service.label}
            </h2>
          </div>

          <p className="mt-2 break-all text-sm text-slate-500">{service.target}</p>
        </div>

        <div
          className={`flex shrink-0 items-center gap-1 rounded-md px-2 py-1 text-xs font-medium ${stateView.badgeClassName}`}
        >
          <stateView.Icon className="h-3.5 w-3.5" aria-hidden="true" />
          <span>{stateView.label}</span>
        </div>
      </div>

      <div className="mt-5 grid grid-cols-2 gap-3 text-sm">
        <div>
          <p className="text-slate-500">状态码</p>
          <p className="mt-1 font-medium text-slate-950">
            {service.statusCode ?? "-"}
          </p>
        </div>

        <div>
          <p className="text-slate-500">耗时</p>
          <p className="mt-1 font-medium text-slate-950">
            {typeof service.responseTimeMs === "number"
              ? `${service.responseTimeMs} ms`
              : "-"}
          </p>
        </div>
      </div>

      <div className="mt-5 rounded-md bg-slate-50 p-3">
        <p className="text-sm text-slate-700">{service.message}</p>
        <p className="mt-2 text-xs text-slate-500">
          检查时间：{formatTime(service.checkedAt)}
        </p>
      </div>
    </section>
  );
}

function getStateView(state: ServiceHealth["state"]) {
  if (state === "healthy") {
    return {
      label: "正常",
      Icon: CheckCircle2,
      badgeClassName: "bg-emerald-50 text-emerald-700",
    };
  }

  if (state === "degraded") {
    return {
      label: "降级",
      Icon: AlertTriangle,
      badgeClassName: "bg-amber-50 text-amber-700",
    };
  }

  return {
    label: "异常",
    Icon: XCircle,
    badgeClassName: "bg-rose-50 text-rose-700",
  };
}

function formatTime(value: string) {
  return new Intl.DateTimeFormat("zh-CN", {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  }).format(new Date(value));
}