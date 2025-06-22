package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/meilisearch/meilisearch-go"
	"michiru/config"
	"michiru/models"
)

var (
	client meilisearch.ServiceManager
	once   sync.Once
)

func getMeilisearchClient(cfg config.Config) meilisearch.ServiceManager {
	once.Do(
		func() {
			c, err := meilisearch.Connect(
				cfg.MeilisearchURL, meilisearch.WithAPIKey(cfg.MeilisearchKey),
			)
			if err != nil {
				logger.Fatalf(
					"FATAL: Could not connect to Meilisearch.\n%v", err,
				)
			}
			client = c
		},
	)

	return client
}

// AddAnime adds all supplied models.AnimeDocument into the search index defined by config.IndexName.
func AddAnime(
	ctx context.Context, cfg config.Config, anime []models.AnimeDocument,
) error {
	c := getMeilisearchClient(cfg)
	idx := c.Index(cfg.IndexName)

	logger.Println("Resetting Meilisearch index")

	deleteTask, err := idx.DeleteAllDocumentsWithContext(ctx)
	if err != nil {
		return fmt.Errorf("error creating document deletion task: %w", err)
	}

	res, err := c.WaitForTaskWithContext(
		ctx, deleteTask.TaskUID, cfg.TaskTimeout,
	)
	if err != nil || res.Status != meilisearch.TaskStatusSucceeded {
		return fmt.Errorf(
			"error waiting for document deletion task completion: %w", err,
		)
	}

	logger.Println("Updating Meilisearch index")

	addTask, err := idx.AddDocumentsWithContext(ctx, anime)
	if err != nil {
		return fmt.Errorf("error creating document insertion task: %w", err)
	}

	res, err = c.WaitForTaskWithContext(ctx, addTask.TaskUID, cfg.TaskTimeout)
	if err != nil || res.Status != meilisearch.TaskStatusSucceeded {
		return fmt.Errorf(
			"error waiting for document insertion task completion: %w", err,
		)
	}

	return nil
}

// UpdateMetadata updates metadata for the index defined by config.IndexName with the supplied document.
func UpdateMetadata(
	ctx context.Context, cfg config.Config, meta *models.MetadataDocument,
) error {
	c := getMeilisearchClient(cfg)
	idx := c.Index("index_metadata")

	logger.Println("Updating metadata")

	meta.Id = fmt.Sprintf("%s", cfg.IndexName)
	task, err := idx.AddDocuments([]models.MetadataDocument{*meta})
	if err != nil {
		return fmt.Errorf("error creating metadata insertion task: %w", err)
	}

	res, err := c.WaitForTaskWithContext(ctx, task.TaskUID, cfg.TaskTimeout)
	if err != nil || res.Status != meilisearch.TaskStatusSucceeded {
		return fmt.Errorf(
			"error waiting for metadata insertion task completion: %w", err,
		)
	}

	return nil
}

// InitIndexes sets up required search indexes in Meilisearch.
// If the indexes already exist, this function does nothing.
func InitIndexes(ctx context.Context, cfg config.Config) error {
	c := getMeilisearchClient(cfg)

	_, notExists := c.GetIndex(cfg.IndexName)
	if notExists != nil {
		logger.Println("Creating search index")

		createIndexTask, err := c.CreateIndexWithContext(
			ctx, &meilisearch.IndexConfig{
				Uid:        cfg.IndexName,
				PrimaryKey: "aid",
			},
		)
		if err != nil {
			return fmt.Errorf("error creating index: %w", err)
		}

		res, err := c.WaitForTaskWithContext(
			ctx, createIndexTask.TaskUID, cfg.TaskTimeout,
		)
		if err != nil || res.Status != meilisearch.TaskStatusSucceeded {
			return fmt.Errorf("error waiting for index creation: %w", err)
		}

		logger.Println("Updating search index settings")

		idx := c.Index(cfg.IndexName)
		updateTask, err := idx.UpdateSettingsWithContext(
			ctx, &meilisearch.Settings{
				DisplayedAttributes: []string{
					"aid",
					"mainTitle",
					"officialTitles",
					"shortTitles",
					"synonymousTitles",
					"kanaTitles",
					"cardTitles",
				},
				// Manually defined to enforce attribute sorting order in order of importance
				SearchableAttributes: []string{
					"mainTitle",
					"officialTitles",
					"shortTitles",
					"synonymousTitles",
					"kanaTitles",
					"cardTitles",
				},
				RankingRules: []string{
					"words",
					"exactness",
					"attribute",
					"typo",
					"proximity",
					"sort",
				},
			},
		)
		if err != nil {
			return fmt.Errorf("error updating index settings: %w", err)
		}

		res, err = c.WaitForTaskWithContext(
			ctx, updateTask.TaskUID, cfg.TaskTimeout,
		)
		if err != nil || res.Status != meilisearch.TaskStatusSucceeded {
			return fmt.Errorf("error waiting for settings update: %w", err)
		}
	}

	_, notExists = c.GetIndex("index_metadata")
	if notExists != nil {
		logger.Println("Creating metadata index")

		createIndexTask, err := c.CreateIndexWithContext(
			ctx, &meilisearch.IndexConfig{
				Uid:        "index_metadata",
				PrimaryKey: "id",
			},
		)
		if err != nil {
			return fmt.Errorf("error creating index: %w", err)
		}

		res, err := c.WaitForTaskWithContext(
			ctx, createIndexTask.TaskUID, cfg.TaskTimeout,
		)
		if err != nil || res.Status != meilisearch.TaskStatusSucceeded {
			return fmt.Errorf("error waiting for index creation: %w", err)
		}
	}

	return nil
}

// ResetIndexes deletes ALL indexes in the connected Meilisearch instance.
func ResetIndexes(ctx context.Context, cfg config.Config) error {
	c := getMeilisearchClient(cfg)
	res, _ := c.ListIndexes(
		&meilisearch.IndexesQuery{
			Limit:  20,
			Offset: 0,
		},
	)

	for _, index := range res.Results {
		task, err := c.DeleteIndexWithContext(ctx, index.UID)
		if err != nil {
			return fmt.Errorf("error submitting delete index task: %w", err)
		}

		cTask, err := c.WaitForTaskWithContext(
			ctx, task.TaskUID, cfg.TaskTimeout,
		)
		if err != nil || cTask.Status != meilisearch.TaskStatusSucceeded {
			return fmt.Errorf(
				"error waiting for index deletion task completion: %w", err,
			)
		}
	}

	return nil
}

func SearchAnime(
	cfg config.Config, params *models.QueryParams,
) ([]models.AnimeSearchDocument, int, error) {
	c := getMeilisearchClient(cfg)
	idx := c.Index(cfg.IndexName)

	res, err := idx.Search(
		params.Query, &meilisearch.SearchRequest{
			Offset:                int64(params.Offset),
			Limit:                 int64(params.Limit),
			AttributesToHighlight: []string{"*"},
			HighlightPreTag:       "<span>",
			HighlightPostTag:      "</span>",
			ShowRankingScore:      true,
		},
	)
	if err != nil {
		return nil, 0, err
	}

	b, err := json.Marshal(res.Hits)
	if err != nil {
		return nil, 0, err
	}

	var results []models.AnimeSearchDocument
	err = json.Unmarshal(b, &results)
	if err != nil {
		return nil, 0, err
	}

	return results, int(res.EstimatedTotalHits), nil
}

func GetMetadata(
	ctx context.Context, cfg config.Config,
) (*models.MetadataDocument, error) {
	c := getMeilisearchClient(cfg)
	idx := c.Index("index_metadata")

	var meta models.MetadataDocument
	err := idx.GetDocumentWithContext(ctx, cfg.IndexName, nil, &meta)
	if err != nil {
		if err.(*meilisearch.Error).StatusCode == 404 {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting metadata: %w", err)
	}

	return &meta, nil
}
