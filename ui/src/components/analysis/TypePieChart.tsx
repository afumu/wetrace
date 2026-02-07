import { PieChart, Pie, Cell, ResponsiveContainer, Legend, Tooltip } from 'recharts';
import type { MessageTypeStat } from '@/api';

interface Props {
  data: MessageTypeStat[];
}

const COLORS = ['#ec4899', '#8b5cf6', '#3b82f6', '#10b981', '#f59e0b', '#6b7280'];

const TYPE_NAMES: Record<number, string> = {
  1: '文本',
  3: '图片',
  34: '语音',
  43: '视频',
  47: '表情',
  48: '位置',
  49: '链接',
  10000: '系统',
};

export function TypePieChart({ data }: Props) {
  const chartData = data.map(item => ({
    name: TYPE_NAMES[item.type] || `类型 ${item.type}`,
    value: item.count
  }));

  return (
    <div className="h-[300px] w-full">
      <ResponsiveContainer width="100%" height="100%">
        <PieChart>
          <Pie
            data={chartData}
            cx="50%"
            cy="50%"
            innerRadius={60}
            outerRadius={80}
            paddingAngle={5}
            dataKey="value"
          >
            {chartData.map((_, index) => (
              <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
            ))}
          </Pie>
          <Tooltip 
            contentStyle={{ borderRadius: '8px', border: 'none', boxShadow: '0 4px 12px rgba(0,0,0,0.1)' }}
          />
          <Legend verticalAlign="bottom" height={36}/>
        </PieChart>
      </ResponsiveContainer>
    </div>
  );
}
