// Copyright 2017 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build !nodevstat

#include <devstat.h>
#include <fcntl.h>
#include <libgeom.h>
#include <limits.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <devstat_freebsd.h>


int _get_stats(struct devinfo *info, Stats **stats) {
	struct statinfo current;
	current.dinfo = info;

	if (devstat_getdevs(NULL, &current) == -1) {
		return -1;
	}

	Stats *p = (Stats*)calloc(current.dinfo->numdevs, sizeof(Stats));
	for (int i = 0; i < current.dinfo->numdevs; i++) {
		uint64_t bytes_read, bytes_write, bytes_free, queue_length;
		uint64_t transfers_other, transfers_read, transfers_write, transfers_free;
		long double duration_other, duration_read, duration_write, duration_free, transfers_per_second;
                long double transfers_per_second_read, mb_per_second_read, ms_per_transaction_read;
		long double transfers_per_second_write, mb_per_second_write, ms_per_transaction_write;
		long double transfers_per_second_free, mb_per_second_free, ms_per_transaction_free;
		long double transfers_per_second_other, mb_per_second_other, ms_per_transaction_other;
		long double kb_per_transfer_read, kb_per_transfer_write, kb_per_transfer_free;
		long double busy_time, busy_pct;
		uint64_t blocks;

		strcpy(p[i].device, current.dinfo->devices[i].device_name);
		p[i].unit = current.dinfo->devices[i].unit_number;
		devstat_compute_statistics(&current.dinfo->devices[i],
				NULL,
				1.0,
				DSM_TOTAL_BYTES_READ, &bytes_read,
				DSM_TOTAL_BYTES_WRITE, &bytes_write,
				DSM_TOTAL_BYTES_FREE, &bytes_free,
				DSM_TOTAL_TRANSFERS_OTHER, &transfers_other,
				DSM_TOTAL_TRANSFERS_READ, &transfers_read,
				DSM_TOTAL_TRANSFERS_WRITE, &transfers_write,
				DSM_TOTAL_TRANSFERS_FREE, &transfers_free,
				DSM_TOTAL_DURATION_OTHER, &duration_other,
				DSM_TOTAL_DURATION_READ, &duration_read,
				DSM_TOTAL_DURATION_WRITE, &duration_write,
				DSM_TOTAL_DURATION_FREE, &duration_free,
				DSM_TOTAL_BUSY_TIME, &busy_time,
				DSM_TOTAL_BLOCKS, &blocks,
				DSM_QUEUE_LENGTH, &queue_length,
				DSM_TRANSFERS_PER_SECOND, &transfers_per_second,
				DSM_TRANSFERS_PER_SECOND_READ, &transfers_per_second_read,
				DSM_MB_PER_SECOND_READ, &mb_per_second_read,
				DSM_MS_PER_TRANSACTION_READ, &ms_per_transaction_read,
				DSM_TRANSFERS_PER_SECOND_WRITE, &transfers_per_second_write,
                                DSM_MB_PER_SECOND_WRITE, &mb_per_second_write,
                                DSM_MS_PER_TRANSACTION_WRITE, &ms_per_transaction_write,
				DSM_BUSY_PCT, busy_pct,
				DSM_TRANSFERS_PER_SECOND_FREE, &transfers_per_second_free,
                                DSM_MB_PER_SECOND_FREE, &mb_per_second_free,
                                DSM_MS_PER_TRANSACTION_FREE, &ms_per_transaction_free,
				DSM_TRANSFERS_PER_SECOND_OTHER, &transfers_per_second_other,
                                DSM_MS_PER_TRANSACTION_OTHER, &ms_per_transaction_other,
				DSM_KB_PER_TRANSFER_READ, &kb_per_transfer_read,
				DSM_KB_PER_TRANSFER_WRITE, &kb_per_transfer_write,
				DSM_KB_PER_TRANSFER_FREE, &kb_per_transfer_free,
				DSM_NONE);

		p[i].bytes.read = bytes_read;
		p[i].bytes.write = bytes_write;
		p[i].bytes.free = bytes_free;
		p[i].transfers.other = transfers_other;
		p[i].transfers.read = transfers_read;
		p[i].transfers.write = transfers_write;
		p[i].transfers.free = transfers_free;
		p[i].duration.other = duration_other;
		p[i].duration.read = duration_read;
		p[i].duration.write = duration_write;
		p[i].duration.free = duration_free;
		p[i].busy_time = busy_time;
		p[i].busy_percent = busy_pct;
		p[i].blocks = blocks;
		p[i].queue_length = queue_length;
		p[i].tps.total = transfers_per_second;
		p[i].tps.read = transfers_per_second_read;
		p[i].tps.write = transfers_per_second_write;
		p[i].tps.free = transfers_per_second_free;
		p[i].tps.other = transfers_per_second_other;
		p[i].mbps.read = mb_per_second_read;
		p[i].mbps.write = mb_per_second_write;
		p[i].kbpt.read = kb_per_transfer_read;
		p[i].kbpt.write = kb_per_transfer_write;
		p[i].kbpt.free = kb_per_transfer_free;
		p[i].mspertxn.read = ms_per_transaction_read;
		p[i].mspertxn.write = ms_per_transaction_write;
		p[i].mspertxn.other = ms_per_transaction_other;
	}

	*stats = p;
	return current.dinfo->numdevs;
}
