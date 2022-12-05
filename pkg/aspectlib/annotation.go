package aspectlib

type (
	Annotation    string
	AnnotationKey string
)

func (a Annotation) String() string { return string(a) }

const (
	// CommentFactory custom annotation declaration
	CommentCustom = Annotation("@Custom")
	// CommentProxy for struct while comment @Proxy generate a file with _gen.go suffix
	CommentProxy = Annotation("@Proxy")
	// CommentPointcut for struct function while comment @Pointcut generate a proxy func for proxy struct
	CommentPointcut = Annotation("@Pointcut")
	// CommentAspect for struct while comment @Aspect then use to enhance other function
	CommentAspect = Annotation("@Aspect")
	// CommentAdviceBefore for struct function while comment @Before then use to enhance other function
	CommentAdviceBefore = Annotation("@Before")
	// CommentAdviceAfter for struct function while comment @After then use to enhance other function
	CommentAdviceAfter = Annotation("@After")
	// CommentAdviceAround for struct function while comment @Around then use to enhance other function
	CommentAdviceAround = Annotation("@Around")

	// CommentKeyDefault key for comment params separated by "="
	CommentKeyDefault = AnnotationKey("")
	// CommentKeyDefault abstract key for @Proxy comment
	CommentKeyAbstract = AnnotationKey("abstract")
	CommentKeySuffix   = AnnotationKey("suffix")
)

var (
	adviceAnnotationList = []Annotation{CommentAdviceBefore, CommentAdviceAfter, CommentAdviceAround}
	funcAnnotationList   = append(adviceAnnotationList, CommentPointcut)
	allCommentKey        = map[AnnotationKey]struct{}{
		CommentKeyDefault:  {},
		CommentKeyAbstract: {},
		CommentKeySuffix:   {},
	}
)
