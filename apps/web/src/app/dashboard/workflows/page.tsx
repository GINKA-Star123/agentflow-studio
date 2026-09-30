import { Badge } from "@/components/ui/badge";
import { WorkflowDesignerLayout } from "@/components/workflow/workflow-designer-layout";

export default function WorkflowsPage() {
  return (
    <div className="space-y-4">
      <header className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-semibold text-slate-950">
            Workflow Designer
          </h1>
          <p className="mt-1 text-sm text-slate-500">可视化编排工作区</p>
        </div>

        <Badge variant="secondary">Draft</Badge>
      </header>

      <WorkflowDesignerLayout />
    </div>
  );
}