import type { MemberActivity } from "@/api";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";

interface Props {
  data: MemberActivity[];
}

export function MemberRank({ data }: Props) {
  const maxCount = data[0]?.messageCount || 1;

  return (
    <div className="space-y-4">
      {data.map((member, index) => (
        <div key={member.platformId} className="flex items-center gap-3">
          <div className="w-6 text-sm font-bold text-muted-foreground">
            {index + 1}
          </div>
          <Avatar className="h-8 w-8">
            <AvatarImage src={member.avatar} />
            <AvatarFallback>{member.name.slice(0, 1)}</AvatarFallback>
          </Avatar>
          <div className="flex-1 min-w-0">
            <div className="flex justify-between mb-1">
              <span className="text-sm font-medium truncate">{member.name}</span>
              <span className="text-sm text-muted-foreground">{member.messageCount}</span>
            </div>
            <div className="h-2 w-full bg-muted rounded-full overflow-hidden">
              <div 
                className="h-full bg-pink-500 rounded-full" 
                style={{ width: `${(member.messageCount / maxCount) * 100}%` }}
              />
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}
