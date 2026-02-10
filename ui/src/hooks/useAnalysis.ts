import { useQuery } from "@tanstack/react-query";
import { analysisApi } from "@/api";

export function useAnalysis(talker: string) {
  const hourlyQuery = useQuery({
    queryKey: ["analysis", "hourly", talker],
    queryFn: () => analysisApi.getHourly(talker),
    enabled: !!talker,
    retry: 1,
  });

  const dailyQuery = useQuery({
    queryKey: ["analysis", "daily", talker],
    queryFn: () => analysisApi.getDaily(talker),
    enabled: !!talker,
    retry: 1,
  });

  const typeQuery = useQuery({
    queryKey: ["analysis", "types", talker],
    queryFn: () => analysisApi.getTypeDistribution(talker),
    enabled: !!talker,
    retry: 1,
  });

  const memberQuery = useQuery({
    queryKey: ["analysis", "members", talker],
    queryFn: () => analysisApi.getMemberActivity(talker),
    enabled: !!talker,
    retry: 1,
  });

  const repeatQuery = useQuery({
    queryKey: ["analysis", "repeat", talker],
    queryFn: () => analysisApi.getRepeat(talker),
    enabled: !!talker,
    retry: 1,
  });

  // 只要最重要的几个查询在加载，就显示全局 loading
  const isInitialLoading = (hourlyQuery.isLoading && hourlyQuery.isFetching) || 
                           (dailyQuery.isLoading && dailyQuery.isFetching);

    return {

      hourly: hourlyQuery,

      daily: dailyQuery,

      types: typeQuery,

      members: memberQuery,

      repeat: repeatQuery,

      isLoading: isInitialLoading,

    };

  }
